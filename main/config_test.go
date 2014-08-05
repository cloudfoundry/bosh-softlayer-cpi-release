package main_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "bosh/system/fakes"

	bslcaction "github.com/maximilien/bosh-softlayer-cpi/action"
	. "github.com/maximilien/bosh-softlayer-cpi/main"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var validConfig = Config{
	Warden:  validWardenConfig,
	Actions: validActionsOptions,
}

var validWardenConfig = WardenConfig{
	ConnectNetwork: "fake-tcp",
	ConnectAddress: "fake-address",
}

var validActionsOptions = bslcaction.ConcreteFactoryOptions{
	StemcellsDir: "/tmp/stemcells",
	DisksDir:     "/tmp/disks",

	HostEphemeralBindMountsDir:  "/tmp/host-ephemeral-bind-mounts-dir",
	HostPersistentBindMountsDir: "/tmp/host-persistent-bind-mounts-dir",

	GuestEphemeralBindMountPath:  "/tmp/guest-ephemeral-bind-mount-path",
	GuestPersistentBindMountsDir: "/tmp/guest-persistent-bind-mounts-dir",

	Agent: bslcvm.AgentOptions{
		Mbus: "fake-mbus",
		NTP:  []string{},

		Blobstore: bslcvm.BlobstoreOptions{
			Type: "fake-blobstore-type",
		},
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

		It("does not return error if all warden and agent sections are valid", func() {
			err := config.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if warden section is not valid", func() {
			config.Warden.ConnectNetwork = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Warden configuration"))
		})

		It("returns error if actions section is not valid", func() {
			config.Actions.DisksDir = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Actions configuration"))
		})
	})
})

var _ = Describe("WardenConfig", func() {
	var (
		config WardenConfig
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			config = validWardenConfig
		})

		It("does not return error if all fields are valid", func() {
			err := config.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if ConnectNetwork is empty", func() {
			config.ConnectNetwork = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty ConnectNetwork"))
		})

		It("returns error if ConnectAddress is empty", func() {
			config.ConnectAddress = ""

			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty ConnectAddress"))
		})
	})
})
