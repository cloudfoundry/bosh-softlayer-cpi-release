package main_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	. "github.com/maximilien/bosh-softlayer-cpi/main"

	bslcaction "github.com/maximilien/bosh-softlayer-cpi/action"
)

var validSoftLayerConfig = SoftLayerConfig{
	Username: "fake-username",
	ApiKey:   "fake-api-key",
}

var validActionsOptions = bslcaction.ConcreteFactoryOptions{
	StemcellsDir: "/tmp/stemcells",
}

var validConfig = Config{
	SoftLayer: validSoftLayerConfig,
	Actions:   validActionsOptions,
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

		_, err = NewConfigFromPath("/config.json", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Validating config"))
	})

	It("returns error if file contains invalid json", func() {
		err := fs.WriteFileString("/config.json", "-")
		Expect(err).ToNot(HaveOccurred())

		_, err = NewConfigFromPath("/config.json", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unmarshalling config"))
	})

	It("returns error if file cannot be read", func() {
		err := fs.WriteFileString("/config.json", "{}")
		Expect(err).ToNot(HaveOccurred())

		fs.ReadFileError = errors.New("fake-read-err")

		_, err = NewConfigFromPath("/config.json", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-read-err"))
	})
})

var _ = Describe("Config", func() {
	var (
		config Config
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
			config.SoftLayer.Username = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating SoftLayer configuration"))
		})
	})
})

var _ = Describe("SoftLayerConfig", func() {
	var (
		config SoftLayerConfig
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			config = validSoftLayerConfig
		})

		It("does not return error if all fields are valid", func() {
			err := config.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if Username is empty", func() {
			config.Username = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty Username"))
		})

		It("returns error if ApiKey is empty", func() {
			config.ApiKey = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty ApiKey"))
		})
	})
})
