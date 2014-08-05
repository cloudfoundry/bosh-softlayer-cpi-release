package vm_test

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	boshlog "bosh/logger"
	fakewrdnclient "github.com/cloudfoundry-incubator/garden/client/fake_warden_client"
	wrdn "github.com/cloudfoundry-incubator/garden/warden"
	fakewrdn "github.com/cloudfoundry-incubator/garden/warden/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/vm"
)

var _ = Describe("WardenAgentEnvService", func() {
	var (
		wardenClient    *fakewrdnclient.FakeClient
		logger          boshlog.Logger
		agentEnvService WardenAgentEnvService
	)

	BeforeEach(func() {
		wardenClient = fakewrdnclient.New()

		wardenClient.Connection.CreateReturns("fake-vm-id", nil)

		containerSpec := wrdn.ContainerSpec{
			Handle:     "fake-vm-id",
			RootFSPath: "fake-root-fs-path",
		}

		container, err := wardenClient.Create(containerSpec)
		Expect(err).ToNot(HaveOccurred())

		logger = boshlog.NewLogger(boshlog.LevelNone)
		agentEnvService = NewWardenAgentEnvService(container, logger)
	})

	Describe("Fetch", func() {
		var (
			runProcess  *fakewrdn.FakeProcess
			outAgentEnv AgentEnv
		)

		BeforeEach(func() {
			runProcess = &fakewrdn.FakeProcess{}
			runProcess.WaitReturns(0, nil)
			wardenClient.Connection.RunReturns(runProcess, nil)
		})

		makeValidAgentEnvTar := func(agentEnv AgentEnv) io.ReadCloser {
			tarBytes := &bytes.Buffer{}

			tarWriter := tar.NewWriter(tarBytes)

			jsonBytes, err := json.Marshal(agentEnv)
			Expect(err).ToNot(HaveOccurred())

			fileHeader := &tar.Header{
				Name: "warden-cpi-agent-env.json",
				Size: int64(len(jsonBytes)),
			}

			err = tarWriter.WriteHeader(fileHeader)
			Expect(err).ToNot(HaveOccurred())

			_, err = tarWriter.Write(jsonBytes)
			Expect(err).ToNot(HaveOccurred())

			err = tarWriter.Close()
			Expect(err).ToNot(HaveOccurred())

			return ioutil.NopCloser(tarBytes)
		}

		BeforeEach(func() {
			outAgentEnv = AgentEnv{AgentID: "fake-agent-id"}
			wardenClient.Connection.StreamOutReturns(makeValidAgentEnvTar(outAgentEnv), nil)
		})

		It("copies agent env into temporary location", func() {
			_, err := agentEnvService.Fetch()
			Expect(err).ToNot(HaveOccurred())

			count := wardenClient.Connection.RunCallCount()
			Expect(count).To(Equal(1))

			expectedProcessSpec := wrdn.ProcessSpec{
				Path: "bash",
				Args: []string{"-c", "cp /var/vcap/bosh/warden-cpi-agent-env.json /tmp/warden-cpi-agent-env.json && chown vcap:vcap /tmp/warden-cpi-agent-env.json"},

				Privileged: true,
			}

			handle, processSpec, processIO := wardenClient.Connection.RunArgsForCall(0)
			Expect(handle).To(Equal("fake-vm-id"))
			Expect(processSpec).To(Equal(expectedProcessSpec))
			Expect(processIO).To(Equal(wrdn.ProcessIO{}))
		})

		Context("when copying agent env into temporary location succeeds", func() {
			Context("when container succeeds to stream out agent env", func() {
				It("returns agent env from temporary location in the container", func() {
					agentEnv, err := agentEnvService.Fetch()
					Expect(err).ToNot(HaveOccurred())
					Expect(agentEnv).To(Equal(outAgentEnv))

					count := wardenClient.Connection.StreamOutCallCount()
					Expect(count).To(Equal(1))

					handle, srcPath := wardenClient.Connection.StreamOutArgsForCall(0)
					Expect(handle).To(Equal("fake-vm-id"))
					Expect(srcPath).To(Equal("/tmp/warden-cpi-agent-env.json"))
				})
			})

			Context("when container fails to stream out because tar stream contains bad header", func() {
				BeforeEach(func() {
					wardenClient.Connection.StreamOutReturns(ioutil.NopCloser(&bytes.Buffer{}), nil)
				})

				It("returns error", func() {
					agentEnv, err := agentEnvService.Fetch()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading tar header for agent env"))
					Expect(agentEnv).To(Equal(AgentEnv{}))
				})
			})

			Context("when container fails to stream out because agent env cannot be deserialized", func() {
				BeforeEach(func() {
					tarBytes := &bytes.Buffer{}

					tarWriter := tar.NewWriter(tarBytes)

					err := tarWriter.WriteHeader(&tar.Header{Name: "warden-cpi-agent-env.json"})
					Expect(err).ToNot(HaveOccurred())

					err = tarWriter.Close()
					Expect(err).ToNot(HaveOccurred())

					wardenClient.Connection.StreamOutReturns(ioutil.NopCloser(tarBytes), nil)
				})

				It("returns error", func() {
					agentEnv, err := agentEnvService.Fetch()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading agent env from tar"))
					Expect(agentEnv).To(Equal(AgentEnv{}))
				})
			})

			Context("when container fails to stream out", func() {
				BeforeEach(func() {
					wardenClient.Connection.StreamOutReturns(nil, errors.New("fake-stream-out-err"))
				})

				It("returns error", func() {
					agentEnv, err := agentEnvService.Fetch()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-stream-out-err"))
					Expect(agentEnv).To(Equal(AgentEnv{}))
				})
			})
		})

		Context("when copying agent env into temporary location fails because command exits with non-0 code", func() {
			BeforeEach(func() {
				runProcess.WaitReturns(1, nil)
			})

			It("returns error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Script exited with non-0 exit code"))
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})

		Context("when copying agent env into temporary location fails", func() {
			BeforeEach(func() {
				runProcess.WaitReturns(0, errors.New("fake-wait-err"))
			})

			It("returns error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-wait-err"))
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})

		Context("when copying agent env into temporary location cannot start", func() {
			BeforeEach(func() {
				wardenClient.Connection.RunReturns(nil, errors.New("fake-run-err"))
			})

			It("returns error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-err"))
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})
	})

	Describe("Update", func() {
		var (
			newAgentEnv AgentEnv
			runProcess  *fakewrdn.FakeProcess
		)

		BeforeEach(func() {
			newAgentEnv = AgentEnv{AgentID: "fake-agent-id"}
		})

		BeforeEach(func() {
			runProcess = &fakewrdn.FakeProcess{}
			runProcess.WaitReturns(0, nil)
			wardenClient.Connection.RunReturns(runProcess, nil)
		})

		It("places infrastructure settings into the container at /tmp/warden-cpi-agent-env.json", func() {
			err := agentEnvService.Update(newAgentEnv)
			Expect(err).ToNot(HaveOccurred())

			count := wardenClient.Connection.StreamInCallCount()
			Expect(count).To(Equal(1))

			handle, dstPath, reader := wardenClient.Connection.StreamInArgsForCall(0)
			Expect(handle).To(Equal("fake-vm-id"))
			Expect(dstPath).To(Equal("/tmp/"))

			tarStream := tar.NewReader(reader)

			header, err := tarStream.Next()
			Expect(err).ToNot(HaveOccurred())
			Expect(header.Name).To(Equal("warden-cpi-agent-env.json")) // todo more?

			jsonBytes := make([]byte, header.Size)

			_, err = tarStream.Read(jsonBytes)
			Expect(err).ToNot(HaveOccurred())

			outAgentEnv, err := NewAgentEnvFromJSON(jsonBytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(outAgentEnv).To(Equal(AgentEnv{AgentID: "fake-agent-id"}))

			_, err = tarStream.Next()
			Expect(err).To(HaveOccurred())
		})

		Context("when streaming into the container succeeds", func() {
			It("moves agent env into final location", func() {
				err := agentEnvService.Update(newAgentEnv)
				Expect(err).ToNot(HaveOccurred())

				count := wardenClient.Connection.RunCallCount()
				Expect(count).To(Equal(1))

				expectedProcessSpec := wrdn.ProcessSpec{
					Path: "bash",
					Args: []string{"-c", "mv /tmp/warden-cpi-agent-env.json /var/vcap/bosh/warden-cpi-agent-env.json"},

					Privileged: true,
				}

				handle, processSpec, processIO := wardenClient.Connection.RunArgsForCall(0)
				Expect(handle).To(Equal("fake-vm-id"))
				Expect(processSpec).To(Equal(expectedProcessSpec))
				Expect(processIO).To(Equal(wrdn.ProcessIO{}))
			})

			Context("when moving agent env into final location fails because command exits with non-0 code", func() {
				BeforeEach(func() {
					runProcess.WaitReturns(1, nil)
				})

				It("returns error", func() {
					err := agentEnvService.Update(newAgentEnv)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Script exited with non-0 exit code"))
				})
			})

			Context("when moving agent env into final location fails", func() {
				BeforeEach(func() {
					runProcess.WaitReturns(0, errors.New("fake-wait-err"))
				})

				It("returns error", func() {
					err := agentEnvService.Update(newAgentEnv)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-wait-err"))
				})
			})

			Context("when moving agent env into final location cannot start", func() {
				BeforeEach(func() {
					wardenClient.Connection.RunReturns(nil, errors.New("fake-run-err"))
				})

				It("returns error", func() {
					agentEnv, err := agentEnvService.Fetch()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-run-err"))
					Expect(agentEnv).To(Equal(AgentEnv{}))
				})
			})
		})

		Context("when container fails to stream in because agent env cannot be serialized", func() {
			BeforeEach(func() {
				newAgentEnv.Env = EnvSpec{"fake-net-name": NonJSONMarshable{}} // cheating!
			})

			It("returns error", func() {
				err := agentEnvService.Update(newAgentEnv)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-marshal-err"))
			})
		})

		Context("when container fails to stream in", func() {
			BeforeEach(func() {
				wardenClient.Connection.StreamInReturns(errors.New("fake-stream-in-err"))
			})

			It("returns error", func() {
				err := agentEnvService.Update(newAgentEnv)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stream-in-err"))
			})
		})
	})
})
