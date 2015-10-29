package vm_test

import (
	"encoding/json"
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakebslvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("SoftlayerAgentEnvService", func() {
	var (
		fakeSoftlayerFileService *fakebslvm.FakeSoftlayerFileService
		agentEnvService          AgentEnvService
	)

	BeforeEach(func() {
		fakeSoftlayerFileService = fakebslvm.NewFakeSoftlayerFileService()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		agentEnvService = NewFSAgentEnvService(fakeSoftlayerFileService, logger)
	})

	Describe("Fetch", func() {
		It("downloads file contents from warden container", func() {
			expectedAgentEnv := AgentEnv{
				AgentID: "fake-agent-id",
			}
			downloadAgentEnvBytes, err := json.Marshal(expectedAgentEnv)
			Expect(err).ToNot(HaveOccurred())

			fakeSoftlayerFileService.DownloadContents = downloadAgentEnvBytes

			agentEnv, err := agentEnvService.Fetch()
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv).To(Equal(expectedAgentEnv))
			Expect(fakeSoftlayerFileService.DownloadSourcePath).To(Equal("/var/vcap/bosh/user_data.json"))
		})

		Context("when container fails to stream out because agent env cannot be deserialized", func() {
			BeforeEach(func() {
				fakeSoftlayerFileService.DownloadContents = []byte("invalid-json")
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
				fakeSoftlayerFileService.DownloadErr = errors.New("fake-download-error")
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
			Expect(fakeSoftlayerFileService.UploadInputs[0].Contents).To(Equal(expectedAgentEnvBytes))
		})

	})
})
