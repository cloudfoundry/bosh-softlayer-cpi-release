package action_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("ConcreteFactoryOptions", func() {
	var (
		options ConcreteFactoryOptions

		validOptions = ConcreteFactoryOptions{

			Agent: bslcvm.AgentOptions{
				Mbus: "fake-mbus",
				NTP:  []string{},

				Blobstore: bslcvm.BlobstoreOptions{
					Provider: "fake-blobstore-type",
				},
			},
			Softlayer: SoftLayerConfig{
				Username: "fake-username",
				ApiKey:   "fke-apikey",
				FeatureOptions: bslcvm.FeatureOptions{
					ApiEndpoint:                      "fake-api-endpoint",
					ApiWaitTime:                      3,
					ApiRetryCount:                    5,
					CreateISCSIVolumePollingInterval: 1,
				},
			},
		}
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			options = validOptions
		})

		It("returns error if agent section is not valid", func() {
			options.Agent.Mbus = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Agent configuration"))
		})

		It("sets the environment variable correctly if it is specified", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())

			Expect(os.Getenv("SL_API_ENDPOINT")).To(Equal("fake-api-endpoint"))
			Expect(os.Getenv("SL_API_WAIT_TIME")).To(Equal("3"))
			Expect(os.Getenv("SL_API_RETRY_COUNT")).To(Equal("5"))
			Expect(os.Getenv("SL_CREATE_ISCSI_VOLUME_POLLING_INTERVAL")).To(Equal("1"))
		})

		It("sets an empty string to the environment variable if it is not specified", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())

			Expect(os.Getenv("SL_CREATE_ISCSI_VOLUME_TIMEOUT")).To(Equal("0"))
		})
	})
})
