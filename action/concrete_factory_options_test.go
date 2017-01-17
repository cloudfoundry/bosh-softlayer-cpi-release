package action_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

var _ = Describe("ConcreteFactoryOptions", func() {
	var (
		options ConcreteFactoryOptions

		validOptions = ConcreteFactoryOptions{

			Agent: AgentOptions{
				Mbus: "fake-mbus",
				NTP:  []string{},

				Blobstore: BlobstoreOptions{
					Provider: "fake-blobstore-type",
				},
			},
			Softlayer: SoftLayerConfig{
				Username:       "fake-username",
				ApiKey:         "fke-apikey",
				FeatureOptions: FeatureOptions{},
			},
		}
	)

	Context("when the option values are specified", func() {
		BeforeEach(func() {
			validOptions.Softlayer.FeatureOptions = FeatureOptions{
				ApiEndpoint:                      "api.service.softlayer.com",
				ApiWaitTime:                      3,
				ApiRetryCount:                    5,
				CreateISCSIVolumeTimeout:         1200,
				CreateISCSIVolumePollingInterval: 20,
			}
			options = validOptions
		})

		It("sets environment variables correctly if specified", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())

			Expect(os.Getenv("SL_API_ENDPOINT")).To(Equal("api.service.softlayer.com"))
			Expect(os.Getenv("SL_API_WAIT_TIME")).To(Equal("3"))
			Expect(os.Getenv("SL_API_RETRY_COUNT")).To(Equal("5"))
			Expect(os.Getenv("SL_CREATE_ISCSI_VOLUME_TIMEOUT")).To(Equal("1200"))
			Expect(os.Getenv("SL_CREATE_ISCSI_VOLUME_POLLING_INTERVAL")).To(Equal("20"))
		})
	})

	Context("when the option values are not specified", func() {
		BeforeEach(func() {
			validOptions.Softlayer.FeatureOptions = FeatureOptions{}
			options = validOptions
		})

		It("returns error if agent section is not valid", func() {
			options.Agent.Mbus = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Agent configuration"))
		})

		It("sets the default values to the environment variables if not specified", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())

			Expect(os.Getenv("SL_API_ENDPOINT")).To(Equal("api.softlayer.com"))
			Expect(os.Getenv("SL_API_WAIT_TIME")).To(Equal("0"))
			Expect(os.Getenv("SL_API_RETRY_COUNT")).To(Equal("1"))
			Expect(os.Getenv("SL_CREATE_ISCSI_VOLUME_TIMEOUT")).To(Equal("600"))
			Expect(os.Getenv("SL_CREATE_ISCSI_VOLUME_POLLING_INTERVAL")).To(Equal("10"))
		})
	})
})
