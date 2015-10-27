package vm_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	wrdn "github.com/cloudfoundry-incubator/garden"
	wrdnclient "github.com/cloudfoundry-incubator/garden/client"
	fakewrdnconn "github.com/cloudfoundry-incubator/garden/client/connection/fakes"
	fakewrdn "github.com/cloudfoundry-incubator/garden/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cppforlife/bosh-warden-cpi/vm"
)

var _ = Describe("WardenFileService", func() {
	var (
		wardenConn   *fakewrdnconn.FakeConnection
		wardenClient wrdnclient.Client

		wardenFileService WardenFileService
	)

	BeforeEach(func() {
		wardenConn = &fakewrdnconn.FakeConnection{}
		wardenClient = wrdnclient.New(wardenConn)

		wardenConn.CreateReturns("fake-vm-id", nil)

		containerSpec := wrdn.ContainerSpec{
			Handle:     "fake-vm-id",
			RootFSPath: "fake-root-fs-path",
		}

		container, err := wardenClient.Create(containerSpec)
		Expect(err).ToNot(HaveOccurred())

		logger := boshlog.NewLogger(boshlog.LevelNone)
		wardenFileService = NewWardenFileService(container, logger)
	})

	Describe("Upload", func() {
		var (
			runProcess *fakewrdn.FakeProcess
		)

		BeforeEach(func() {
			runProcess = &fakewrdn.FakeProcess{}
			runProcess.WaitReturns(0, nil)
			wardenConn.RunReturns(runProcess, nil)
		})

		It("places content into the container at /tmp/destination", func() {
			err := wardenFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
			Expect(err).ToNot(HaveOccurred())

			count := wardenConn.StreamInCallCount()
			Expect(count).To(Equal(1))

			handle, spec := wardenConn.StreamInArgsForCall(0)
			Expect(handle).To(Equal("fake-vm-id"))
			Expect(spec.Path).To(Equal("/tmp/"))
			Expect(spec.User).To(Equal("root"))

			tarStream := tar.NewReader(spec.TarStream)

			header, err := tarStream.Next()
			Expect(err).ToNot(HaveOccurred())
			Expect(header.Name).To(Equal("file.ext")) // todo more?

			contentBytes := make([]byte, header.Size)

			_, err = tarStream.Read(contentBytes)
			Expect(err).ToNot(HaveOccurred())

			Expect(contentBytes).To(Equal([]byte("fake-contents")))

			_, err = tarStream.Next()
			Expect(err).To(HaveOccurred())
		})

		Context("when streaming into the container succeeds", func() {
			It("moves the temporary file into the final location", func() {
				err := wardenFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
				Expect(err).ToNot(HaveOccurred())

				count := wardenConn.RunCallCount()
				Expect(count).To(Equal(1))

				expectedProcessSpec := wrdn.ProcessSpec{
					Path: "bash",
					Args: []string{"-c", "mv /tmp/file.ext /var/vcap/file.ext"},
					User: "root",
				}

				handle, processSpec, _ := wardenConn.RunArgsForCall(0)
				Expect(handle).To(Equal("fake-vm-id"))
				Expect(processSpec).To(Equal(expectedProcessSpec))
			})

			Context("when moving the temporary file into the final location fails because command exits with non-0 code", func() {
				BeforeEach(func() {
					runProcess.WaitReturns(1, nil)
				})

				It("returns error", func() {
					err := wardenFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Script exited with non-0 exit code"))
				})
			})

			Context("when moving the temporary file into the final location fails", func() {
				BeforeEach(func() {
					runProcess.WaitReturns(0, errors.New("fake-wait-err"))
				})

				It("returns error", func() {
					err := wardenFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-wait-err"))
				})
			})

			Context("when moving the temporary file into the final location cannot start", func() {
				BeforeEach(func() {
					wardenConn.RunReturns(nil, errors.New("fake-run-err"))
				})

				It("returns error", func() {
					err := wardenFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-run-err"))
				})
			})
		})

		Context("when container fails to stream in", func() {
			BeforeEach(func() {
				wardenConn.StreamInReturns(errors.New("fake-stream-in-err"))
			})

			It("returns error", func() {
				err := wardenFileService.Upload("/var/vcap/file.ext", []byte("fake-contents"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stream-in-err"))
			})
		})
	})

	Describe("Download", func() {
		var (
			runProcess *fakewrdn.FakeProcess
		)

		BeforeEach(func() {
			runProcess = &fakewrdn.FakeProcess{}
			runProcess.WaitReturns(0, nil)
			wardenConn.RunReturns(runProcess, nil)
		})

		makeValidAgentEnvTar := func() io.ReadCloser {
			tarBytes := &bytes.Buffer{}

			tarWriter := tar.NewWriter(tarBytes)

			contents := []byte("fake-contents")

			fileHeader := &tar.Header{
				Name: "warden-cpi-agent-env.json",
				Size: int64(len(contents)),
			}

			err := tarWriter.WriteHeader(fileHeader)
			Expect(err).ToNot(HaveOccurred())

			_, err = tarWriter.Write(contents)
			Expect(err).ToNot(HaveOccurred())

			err = tarWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			return ioutil.NopCloser(tarBytes)
		}

		BeforeEach(func() {
			wardenConn.StreamOutReturns(makeValidAgentEnvTar(), nil)
		})

		It("copies agent env into temporary location", func() {
			_, err := wardenFileService.Download("/fake-download-path/file.ext")
			Expect(err).ToNot(HaveOccurred())

			count := wardenConn.RunCallCount()
			Expect(count).To(Equal(1))

			expectedProcessSpec := wrdn.ProcessSpec{
				Path: "bash",
				Args: []string{"-c", "cp /fake-download-path/file.ext /tmp/file.ext && chown vcap:vcap /tmp/file.ext"},
				User: "root",
			}

			handle, processSpec, _ := wardenConn.RunArgsForCall(0)
			Expect(handle).To(Equal("fake-vm-id"))
			Expect(processSpec).To(Equal(expectedProcessSpec))
		})

		Context("when copying agent env into temporary location succeeds", func() {
			Context("when container succeeds to stream out agent env", func() {
				It("returns agent env from temporary location in the container", func() {
					contents, err := wardenFileService.Download("/fake-download-path/file.ext")
					Expect(err).ToNot(HaveOccurred())
					Expect(contents).To(Equal([]byte("fake-contents")))

					count := wardenConn.StreamOutCallCount()
					Expect(count).To(Equal(1))

					handle, spec := wardenConn.StreamOutArgsForCall(0)
					Expect(handle).To(Equal("fake-vm-id"))
					Expect(spec.Path).To(Equal("/tmp/file.ext"))
					Expect(spec.User).To(Equal("root"))
				})
			})

			Context("when container fails to stream out because tar stream contains bad header", func() {
				BeforeEach(func() {
					wardenConn.StreamOutReturns(ioutil.NopCloser(&bytes.Buffer{}), nil)
				})

				It("returns error", func() {
					contents, err := wardenFileService.Download("/fake-download-path/file.ext")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading tar header for 'file.ext'"))
					Expect(contents).To(Equal([]byte{}))
				})
			})

			Context("when container fails to stream out", func() {
				BeforeEach(func() {
					wardenConn.StreamOutReturns(nil, errors.New("fake-stream-out-err"))
				})

				It("returns error", func() {
					contents, err := wardenFileService.Download("/fake-download-path/file.ext")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-stream-out-err"))
					Expect(contents).To(Equal([]byte{}))
				})
			})
		})

		Context("when copying file into temporary location fails because command exits with non-0 code", func() {
			BeforeEach(func() {
				runProcess.WaitReturns(1, nil)
			})

			It("returns error", func() {
				contents, err := wardenFileService.Download("/fake-download-path/file.ext")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Script exited with non-0 exit code"))
				Expect(contents).To(Equal([]byte{}))
			})
		})

		Context("when copying file into temporary location fails", func() {
			BeforeEach(func() {
				runProcess.WaitReturns(0, errors.New("fake-wait-err"))
			})

			It("returns error", func() {
				contents, err := wardenFileService.Download("/fake-download-path/file.ext")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-wait-err"))
				Expect(contents).To(Equal([]byte{}))
			})
		})

		Context("when copying file into temporary location cannot start", func() {
			BeforeEach(func() {
				wardenConn.RunReturns(nil, errors.New("fake-run-err"))
			})

			It("returns error", func() {
				contents, err := wardenFileService.Download("/fake-download-path/file.ext")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-err"))
				Expect(contents).To(Equal([]byte{}))
			})
		})
	})
})
