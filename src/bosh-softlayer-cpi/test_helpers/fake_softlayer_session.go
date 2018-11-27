package test_helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega/ghttp"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

const SoftLayer_WebService_RateLimitExceeded_Exception = "SoftLayer_Exception_WebService_RateLimitExceeded"

type FakeTransportHandler struct {
	FakeServer           *ghttp.Server
	SoftlayerAPIEndpoint string
	MaxRetries           int
}

type rawString struct {
	Val string
}

func NewFakeSoftlayerSession(transportHandler *FakeTransportHandler) *session.Session {
	return &session.Session{
		TransportHandler: transportHandler,
	}
}

func DestroyServer(server *ghttp.Server) {
	server.Close()
}

func (h FakeTransportHandler) DoRequest(sess *session.Session, service string, method string, args []interface{}, options *sl.Options, pResult interface{}) error {
	var (
		restMethod string

		resp []byte
		code int
		err  error
	)

	restMethod = httpMethod(method, args)

	// Parse any method parameters and determine the HTTP method
	// And inject parameters into request body
	var parameters []byte
	if len(args) > 0 {
		// parse the parameters
		parameters, err = json.Marshal(
			map[string]interface{}{
				"parameters": args,
			})
		if err != nil {
			return err
		}
	}

	// Build request path without querystring
	path := buildPath(service, method, options)

	// Build request querystring
	query := encodeQuery(options)

	//Do request
	for try := 0; try <= h.MaxRetries; try++ {
		resp, code, err = h.makeHTTPRequest(
			path,
			query,
			restMethod,
			bytes.NewBuffer(parameters),
			options,
		)

		if err != nil {
			return sl.Error{Wrapped: err}
		}

		sleep := func() {
			time.Sleep(200 * time.Millisecond << uint64(try*2))
		}

		if code < 200 || code > 299 {
			e := sl.Error{StatusCode: code}

			err = json.Unmarshal(resp, &e)

			// If unparseable, wrap the json error
			if err != nil {
				e.Wrapped = err
				e.Message = err.Error()

				if e.Exception == SoftLayer_WebService_RateLimitExceeded_Exception {
					sleep()
					continue
				}
			}
			return e
		}

		break
	}

	// Some APIs that normally return a collection, omit the []'s when the API returns a single value
	returnType := reflect.TypeOf(pResult).String()
	if strings.Index(returnType, "[]") == 1 && strings.Index(string(resp), "[") != 0 {
		resp = []byte("[" + string(resp) + "]")
	}

	// At this point, all that's left to do is parse the return value to the appropriate type, and return
	// any parse errors (or nil if successful)

	err = nil
	switch pResult.(type) {
	case *[]uint8:
		// exclude quotes
		*pResult.(*[]uint8) = resp[1 : len(resp)-1]
	case *datatypes.Void:
	case *uint:
		var val uint64
		val, err = strconv.ParseUint(string(resp), 0, 64)
		if err == nil {
			*pResult.(*uint) = uint(val)
		}
	case *bool:
		*pResult.(*bool), err = strconv.ParseBool(string(resp))
	case *string:
		str := string(resp)
		strIdx := len(str) - 1
		if str == "null" {
			str = ""
		} else if str[0] == '"' && str[strIdx] == '"' {
			rawStr := rawString{str}
			err = json.Unmarshal([]byte(`{"val":`+str+`}`), &rawStr)
			if err == nil {
				str = rawStr.Val
			}
		}
		*pResult.(*string) = str
	default:
		// Must be a json representation of one of the many softlayer datatypes
		err = json.Unmarshal(resp, pResult)
	}

	if err != nil {
		err = sl.Error{Message: err.Error(), Wrapped: err}
	}

	return err
}

func httpMethod(name string, args []interface{}) string {
	if name == "deleteObject" {
		return "DELETE"
	} else if name == "editObject" || name == "editObjects" {
		return "PUT"
	} else if name == "createObject" || name == "createObjects" || len(args) > 0 {
		return "POST"
	}

	return "GET"
}

func encodeQuery(opts *sl.Options) string {
	query := new(url.URL).Query()

	if opts.Mask != "" {
		query.Add("objectMask", opts.Mask)
	}

	if opts.Filter != "" {
		query.Add("objectFilter", opts.Filter)
	}

	// resultLimit=<offset>,<limit>
	// If offset unspecified, default to 0
	if opts.Limit != nil {
		startOffset := 0
		if opts.Offset != nil {
			startOffset = *opts.Offset
		}

		query.Add("resultLimit", fmt.Sprintf("%d,%d", startOffset, *opts.Limit))
	}

	return query.Encode()
}

func (h FakeTransportHandler) makeHTTPRequest(path string, querystring string, requestType string, requestBody *bytes.Buffer, options *sl.Options) ([]byte, int, error) {
	tr := &http.Transport{DisableKeepAlives: true}
	client := &http.Client{Transport: tr}

	queryUrl := h.SoftlayerAPIEndpoint
	queryUrl = fmt.Sprintf("%s/%s", strings.TrimRight(queryUrl, "/"), path)

	req, err := http.NewRequest(requestType, queryUrl, requestBody)
	if err != nil {
		return nil, 0, err
	}

	req.URL.RawQuery = querystring
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return nil, 520, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return responseBody, resp.StatusCode, nil
}

func buildPath(service string, method string, options *sl.Options) string {
	path := service

	if options.Id != nil {
		path = path + "/" + strconv.Itoa(*options.Id)
	}

	// omit the API method name if the method represents one of the basic REST methods
	if method != "getObject" && method != "deleteObject" && method != "createObject" &&
		method != "createObjects" && method != "editObject" && method != "editObjects" {
		path = path + "/" + method
	}

	return path + ".json"
}
