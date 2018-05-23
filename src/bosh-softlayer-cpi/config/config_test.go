package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//. "bosh-softlayer-cpi/softlayer/common"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	"bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/registry"
	boslconfig "bosh-softlayer-cpi/softlayer/config"
)

var validProperties = config.CPIProperties{
	SoftLayer: validSoftLayerConfig,
	Agent:     validAgentOption,
	Registry:  validClientOptions,
}

var validAgentOption = registry.AgentOptions{
	Mbus:      "fake-mubus",
	Ntp:       []string{""},
	Blobstore: validBlobstoreOptions,
}

var validBlobstoreOptions = registry.BlobstoreOptions{
	Provider: "local",
}

var validSoftLayerConfig = boslconfig.Config{
	Username: "fake-username",
	ApiKey:   "fake-api-key",
}

var validClientOptions = registry.ClientOptions{
	Username: "registry",
	Password: "1330c82d-4bc4-4544-4a90-c2c78fa66431",
	Address:  "127.0.0.1",
	HTTPOptions: registry.HttpRegistryOptions{
		Port:     8000,
		User:     "registry",
		Password: "1330c82d-4bc4-4544-4a90-c2c78fa66431",
	},
	Endpoint: "http://registry:1330c82d-4bc4-4544-4a90-c2c78fa66431@127.0.0.1:8000",
}

var validCloudConfig = config.Cloud{
	Plugin:     "softlayer",
	Properties: validProperties,
}

var validConfig = config.Config{
	Cloud: validCloudConfig,
}

var _ = Describe("NewConfigFromPath", func() {
	var (
		fs *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
	})

	It("returns error if config is not valid", func() {
		err := fs.WriteFileString("/config.json", "{}")
		Expect(err).ToNot(HaveOccurred())

		_, err = config.NewConfigFromPath("/config.json", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Validating config"))
	})

	It("returns error if file contains invalid json", func() {
		err := fs.WriteFileString("/config.json", "-")
		Expect(err).ToNot(HaveOccurred())

		_, err = config.NewConfigFromPath("/config.json", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unmarshalling config"))
	})

	It("returns error if file cannot be read", func() {
		err := fs.WriteFileString("/config.json", "{}")
		Expect(err).ToNot(HaveOccurred())

		fs.ReadFileError = errors.New("fake-read-err")

		_, err = config.NewConfigFromPath("/config.json", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-read-err"))
	})
})

var _ = Describe("Config", func() {
	var (
		config config.Config
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			config = validConfig
		})

		It("does not return error if all softlayer and agent sections are valid", func() {
			err := config.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if softlayer section is not valid", func() {
			config.Cloud.Properties.SoftLayer.Username = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Cloud Properties"))
		})
	})
})
