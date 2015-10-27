package vm_test

import (
	"encoding/json"
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakebwcvm "github.com/cppforlife/bosh-warden-cpi/vm/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-warden-cpi/vm"
)

var _ = Describe("WardenAgentEnvService", func() {
	var (
		fakeWardenFileService *fakebwcvm.FakeWardenFileService
		agentEnvService       AgentEnvService
	)

	BeforeEach(func() {
		fakeWardenFileService = fakebwcvm.NewFakeWardenFileService()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		agentEnvService = NewFSAgentEnvService(fakeWardenFileService, logger)
	})

	Describe("Fetch", func() {
		It("downloads file contents from warden container", func() {
			expectedAgentEnv := AgentEnv{
				AgentID: "fake-agent-id",
			}
			downloadAgentEnvBytes, err := json.Marshal(expectedAgentEnv)
			Expect(err).ToNot(HaveOccurred())

			fakeWardenFileService.DownloadContents = downloadAgentEnvBytes

			agentEnv, err := agentEnvService.Fetch()
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv).To(Equal(expectedAgentEnv))
			Expect(fakeWardenFileService.DownloadSourcePath).To(Equal("/var/vcap/bosh/warden-cpi-agent-env.json"))
		})

		Context("when container fails to stream out because agent env cannot be deserialized", func() {
			BeforeEach(func() {
				fakeWardenFileService.DownloadContents = []byte("invalid-json")
			})

			It("returns error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshalling agent env"))
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})

		Context("when warden fails to download from container", func() {
			BeforeEach(func() {
				fakeWardenFileService.DownloadErr = errors.New("fake-download-error")
			})

			It("returns error", func() {
				agentEnv, err := agentEnvService.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-download-error"))
				Expect(agentEnv).To(Equal(AgentEnv{}))
			})
		})
	})

	Describe("Update", func() {
		var (
			newAgentEnv           AgentEnv
			expectedAgentEnvBytes []byte
		)

		BeforeEach(func() {
			newAgentEnv = AgentEnv{
				AgentID: "fake-agent-id",
			}
			var err error
			expectedAgentEnvBytes, err = json.Marshal(newAgentEnv)
			Expect(err).ToNot(HaveOccurred())
		})

		It("uploads file contents to the warden container", func() {
			err := agentEnvService.Update(newAgentEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeWardenFileService.UploadInputs[0].Contents).To(Equal(expectedAgentEnvBytes))
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

		Context("when warden fails to upload to container", func() {
			BeforeEach(func() {
				fakeWardenFileService.UploadErr = errors.New("fake-upload-error")
			})

			It("returns error", func() {
				err := agentEnvService.Update(newAgentEnv)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-upload-error"))
			})
		})
	})
})
