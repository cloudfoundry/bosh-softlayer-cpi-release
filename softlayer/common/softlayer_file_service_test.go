package common_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakesutil "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

var _ = Describe("SoftlayerFileService", func() {
	var (
		logger               boshlog.Logger
		sshClient            *fakesutil.FakeSshClient
		softlayerFileService SoftlayerFileService
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		sshClient = &fakesutil.FakeSshClient{}
		softlayerFileService = NewSoftlayerFileService(sshClient, logger)
	})

	Describe("Upload", func() {
		It("uploads file contents to the target", func() {
			err := softlayerFileService.Upload("root", "root-password", "fake-backend-ip", "/target/file.ext", []byte("fake-contents"))
			Expect(err).ToNot(HaveOccurred())
			Expect(sshClient.UploadCallCount()).To(Equal(1))

			u, p, a, w, destPath := sshClient.UploadArgsForCall(0)
			buf, ok := w.(*bytes.Buffer)
			Expect(ok).To(BeTrue())

			Expect(u).To(Equal("root"))
			Expect(p).To(Equal("root-password"))
			Expect(a).To(Equal("fake-backend-ip"))
			Expect(destPath).To(Equal("/target/file.ext"))
			Expect(buf.Bytes()).To(Equal([]byte("fake-contents")))
		})

		Context("when upload fails", func() {
			BeforeEach(func() {
				sshClient.UploadReturns(errors.New("boom"))
			})

			It("returns an error", func() {
				err := softlayerFileService.Upload("root", "root-password", "fake-backend-ip", "/target/file.ext", []byte("fake-contents"))
				Expect(err).To(MatchError(`Upload to "/target/file.ext" failed: boom`))
			})
		})
	})

	Describe("Download", func() {
		It("downloads the remote file", func() {
			sshClient.DownloadStub = func(_, _, _, _ string, w io.Writer) error {
				w.Write([]byte("fake-contents"))
				return nil
			}

			contents, err := softlayerFileService.Download("root", "root-password", "fake-backend-ip", "/source/file.ext")
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal([]byte("fake-contents")))

			Expect(sshClient.DownloadCallCount()).To(Equal(1))
			u, p, a, sourcePath, _ := sshClient.DownloadArgsForCall(0)
			Expect(u).To(Equal("root"))
			Expect(p).To(Equal("root-password"))
			Expect(a).To(Equal("fake-backend-ip"))
			Expect(sourcePath).To(Equal("/source/file.ext"))
		})

		Context("when download fails", func() {
			BeforeEach(func() {
				sshClient.DownloadReturns(errors.New("boom"))
			})

			It("returns an error", func() {
				_, err := softlayerFileService.Download("root", "root-password", "fake-backend-ip", "/source/file.ext")
				Expect(err).To(MatchError(`Download of "/source/file.ext" failed: boom`))
			})
		})
	})
})
