package config_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	bslcaction "bosh-softlayer-cpi/action"
	boslconfig "bosh-softlayer-cpi/softlayer/config"

	"bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/registry"
)

var validProperties = bslcaction.ConcreteFactoryOptions{
	Agent:    validAgentOption,
	Registry: validRegistryOptions,
}

var validAgentOption = registry.AgentOptions{
	Mbus:      "fake-mubus",
	Ntp:       []string{""},
	Blobstore: validBlobstoreOptions,
}

var validBlobstoreOptions = registry.BlobstoreOptions{
	Provider: "local",
}

var validRegistryOptions = registry.ClientOptions{
	Protocol: "http",
	Host:     "fake-registry-host",
	Port:     25777,
	Username: "fake-registry-username",
	Password: "fake-registry-password",
	HTTPOptions: registry.HttpRegistryOptions{
		Port:     25777,
		User:     "fake-registry-username",
		Password: "fake-registry-password",
	},
}

var validConfig = config.Config{
	Cloud: validCloudConfig,
}

var validCloudConfig = config.Cloud{
	Plugin: "softlayer",
	Properties: config.CPIProperties{
		SoftLayer: boslconfig.Config{
			Username:             "fake-username",
			ApiKey:               "fake-api-key",
			ApiEndpoint:          "fake-api-endpoint",
			DisableOsReload:      true,
			PublicKey:            "fake-ssh-public-key",
			PublicKeyFingerPrint: "fake-ssh-public-key-fingerprint",
			VpsEndpoint:          "fake-vps-endpoint",
		},
		Agent:    validAgentOption,
		Registry: validRegistryOptions,
	},
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
			Expect(err.Error()).To(ContainSubstring("Validating SoftLayer configuration"))
		})
	})
})
