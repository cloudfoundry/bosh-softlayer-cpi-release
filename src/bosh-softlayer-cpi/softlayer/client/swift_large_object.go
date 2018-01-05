package client

// Refs: https://github.com/fanatic/swift-cli/blob/master/lo.go

import (
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/ncw/swift"
	"hash"
	"io"
	"math"
	"strings"
	"sync"
	"syscall"
	"time"

	cpiLogger "bosh-softlayer-cpi/logger"
)

const (
	swiftLargeObjectLogTag = "SwiftLargeObject"

	// defined by openstack
	minPartSize = 5 * 1024 * 1024
	maxPartSize = 5 * 1024 * 1024 * 1024
)

type part struct {
	r   io.ReadSeeker
	len int64
	b   *bytes.Buffer

	// read by xml encoder
	PartNumber int
	ETag       string

	// Used for checksum of checksums on completion
	contentMd5 string
}

type largeObject struct {
	c          *swift.Connection
	container  string
	objectName string
	timestamp  string
	expire     string

	bufsz      int64
	buf        *bytes.Buffer
	ch         chan *part
	part       int
	closed     bool
	err        error
	wg         sync.WaitGroup
	md5OfParts hash.Hash
	md5        hash.Hash

	bp *bp

	logger cpiLogger.Logger
}

type qBuf struct {
	when   time.Time
	buffer *bytes.Buffer
}

type bp struct {
	makes int
	get   chan *bytes.Buffer
	give  chan *bytes.Buffer
	quit  chan bool
}

func makeBuffer(size int64) []byte {
	return make([]byte, 0, size)
}

func newBufferPool(bufsz int64) (np *bp) {
	np = new(bp)
	np.get = make(chan *bytes.Buffer)
	np.give = make(chan *bytes.Buffer)
	np.quit = make(chan bool)
	go func() {
		q := new(list.List)
		for {
			if q.Len() == 0 {
				size := bufsz + 100*1024 // allocate overhead to avoid slice growth
				q.PushFront(qBuf{when: time.Now(), buffer: bytes.NewBuffer(makeBuffer(int64(size)))})
				np.makes++
			}

			e := q.Front()

			timeout := time.NewTimer(time.Minute)
			select {
			case b := <-np.give:
				timeout.Stop()
				q.PushFront(qBuf{when: time.Now(), buffer: b})

			case np.get <- e.Value.(qBuf).buffer:
				timeout.Stop()
				q.Remove(e)

			case <-timeout.C:
				// free unused buffers
				e := q.Front()
				for e != nil {
					n := e.Next()
					if time.Since(e.Value.(qBuf).when) > time.Minute {
						q.Remove(e)
						e.Value = nil
					}
					e = n
				}
			case <-np.quit:
				return
			}
		}

	}()
	return np
}

// NewLargeObject provides a io.writer to upload data as a segmented upload
//
// It will upload all the segments into a second container named <container>.
// These segments will have names like large_file/1290206778.25/00000000,
// large_file/1290206778.25/00000001, etc.
//
// The main benefit for using a separate container is that the main container listings
// will not be polluted with all the segment names. The reason for using the segment
// name format of <name>/<timestamp>/<segment> is so that an upload of a new
// file with the same name won’t overwrite the contents of the first until the last
// moment when the manifest file is updated.
//
// swift will manage these segment files for you, deleting old segments on deletes
// and overwrites, etc. You can override this behavior with the --leave-segments
// option if desired; this is useful if you want to have multiple versions of
// the same large object available.
func NewLargeObject(c *swift.Connection, path string, concurrency int, partSize int64, expireAfter int64, logger cpiLogger.Logger) (*largeObject, error) {
	pathParts := strings.SplitN(path, "/", 2)
	objectName := "upload"
	if len(pathParts) > 1 {
		objectName = pathParts[1]
	}
	lo := largeObject{
		c:          c,
		container:  pathParts[0],
		objectName: objectName,
		timestamp:  fmt.Sprintf("%d", time.Now().UnixNano()),
		expire:     fmt.Sprintf("%d", expireAfter),

		bufsz: max64(minPartSize, partSize),

		ch:         make(chan *part),
		md5OfParts: md5.New(),
		md5:        md5.New(),

		bp: newBufferPool(minPartSize),

		logger: logger,
	}

	for i := 0; i < max(concurrency, 1); i++ {
		go lo.worker()
	}

	// Create segment container if it doesn't already exist
	err := c.ContainerCreate(lo.container, nil)
	if err != nil {
		return nil, err
	}

	return &lo, nil
}

func (lo *largeObject) Write(b []byte) (int, error) {
	if lo.closed {
		lo.abort()
		return 0, syscall.EINVAL
	}
	if lo.err != nil {
		lo.abort()
		return 0, lo.err
	}
	if lo.buf == nil {
		lo.buf = <-lo.bp.get
		lo.buf.Reset()
	}
	n, err := lo.buf.Write(b)
	if err != nil {
		lo.abort()
		return n, err
	}

	if int64(lo.buf.Len()) >= lo.bufsz {
		lo.flush()
	}
	return n, nil
}

func (lo *largeObject) flush() {
	lo.wg.Add(1)
	lo.part++
	b := *lo.buf
	part := &part{bytes.NewReader(b.Bytes()), int64(b.Len()), lo.buf, lo.part, "", ""}
	var err error
	part.contentMd5, part.ETag, err = lo.md5Content(part.r)
	if err != nil {
		lo.err = err
	}

	lo.ch <- part
	lo.buf = nil
	// double buffer size every 1000 parts to
	// avoid exceeding the 10000-part AWS limit
	// while still reaching the 5 Terabyte max object size
	if lo.part%1000 == 0 {
		lo.bufsz = min64(lo.bufsz*2, maxPartSize)
	}

}

func (lo *largeObject) worker() {
	for part := range lo.ch {
		lo.retryPutPart(part)
	}
}

// Calls putPart up to nTry times to recover from transient errors.
func (lo *largeObject) retryPutPart(part *part) {
	defer lo.wg.Done()
	var err error
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(math.Exp2(float64(i))) * 100 * time.Millisecond) // exponential back-off
		err = lo.putPart(part)
		if err == nil {
			lo.bp.give <- part.b
			return
		}
		lo.logger.Error(swiftLargeObjectLogTag, fmt.Sprintf("Error on attempt %d: Retrying part: %v, Error: %s", i, part, err))
	}
	lo.err = err
}

// uploads a part, checking the etag against the calculated value
func (lo *largeObject) putPart(part *part) error {
	container := lo.container
	objectName := lo.objectName + "/" + lo.timestamp + "/" + fmt.Sprintf("%d", part.PartNumber)

	lo.logger.Debug(swiftLargeObjectLogTag, "Upload Part: (", container, objectName, part.len, fmt.Sprintf("%x", part.contentMd5), part.ETag, ")")

	if _, err := part.r.Seek(0, 0); err != nil { // move back to beginning, if retrying
		return err
	}

	headers, err := lo.c.ObjectPut(container, objectName, part.r, true, "", "", nil)
	if err != nil {
		return err
	}

	s := headers["Etag"]
	if part.ETag != s {
		return fmt.Errorf("Response etag does not match. Remote:%s Calculated: %s", s, part.ETag)
	}
	return nil
}

func (lo *largeObject) Close() (err error) {
	if lo.closed {
		lo.abort()
		return syscall.EINVAL
	}
	if lo.buf != nil {
		buf := *lo.buf
		if buf.Len() > 0 {
			lo.flush()
		}
	}
	lo.wg.Wait()
	close(lo.ch)
	lo.closed = true
	lo.bp.quit <- true

	if lo.part == 0 {
		lo.abort()
		return fmt.Errorf("0 bytes written")
	}
	if lo.err != nil {
		lo.abort()
		return lo.err
	}
	// Complete Multipart upload
	lo.logger.Debug(swiftLargeObjectLogTag, "Complete multipart: (", lo.container, lo.objectName, "X-Object-Manifest: ", lo.container+"/"+lo.objectName+"/"+lo.timestamp, ")")

	reqHeaders := map[string]string{"X-Object-Manifest": lo.container + "/" + lo.objectName + "/" + lo.timestamp}

	if lo.expire != "0" {
		reqHeaders["X-Delete-After"] = lo.expire
	}

	var headers swift.Headers
	for i := 0; i < 3; i++ { //3 retries
		headers, err = lo.c.ObjectPut(lo.container, lo.objectName, strings.NewReader(""), true, "", "", reqHeaders)
		if err == nil {
			break
		}
	}
	if err != nil {
		lo.abort()
		return err
	}
	lo.logger.Debug(swiftLargeObjectLogTag, fmt.Sprintf("Set multipart header: %#v", headers))

	return
}

// Try to abort multipart upload. Do not error on failure.
func (lo *largeObject) abort() {
	objects, err := lo.c.ObjectNamesAll(lo.container, nil)
	if err != nil {
		lo.logger.Error(swiftLargeObjectLogTag, fmt.Sprintf("Return all multipart objects: %v\n", err))
		return
	}
	for _, object := range objects {
		if strings.HasPrefix(object, lo.objectName+"/"+lo.timestamp+"/") {
			lo.c.ObjectDelete(lo.container, object)
			if err != nil {
				lo.logger.Error(swiftLargeObjectLogTag, fmt.Sprintf("Delete the multipart objects: %v\n", err))
			}
		}
	}
	return
}

// Md5 functions
func (lo *largeObject) md5Content(r io.ReadSeeker) (string, string, error) {
	h := md5.New()
	mw := io.MultiWriter(h, lo.md5)
	if _, err := io.Copy(mw, r); err != nil {
		return "", "", err
	}
	sum := h.Sum(nil)
	hexSum := fmt.Sprintf("%x", sum)
	// add to checksum of all parts for verification on upload completion
	if _, err := lo.md5OfParts.Write(sum); err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(sum), hexSum, nil
}

// Put md5 file in .md5 subdirectory of bucket  where the file is stored
// e.g. the md5 for https://mybucket.s3.amazonaws.com/gof3r will be stored in
// https://mybucket.s3.amazonaws.com/gof3r.md5
func (lo *largeObject) putMd5() (err error) {
	calcMd5 := fmt.Sprintf("%x", lo.md5.Sum(nil))
	md5Reader := strings.NewReader(calcMd5)
	lo.logger.Debug(swiftLargeObjectLogTag, fmt.Sprintf("Put md5: %s of object: %s", calcMd5, lo.container+"/"+lo.objectName+".md5"))
	_, err = lo.c.ObjectPut(lo.container, lo.objectName+".md5", md5Reader, true, "", "", nil)
	return
}

// Max functions
func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
