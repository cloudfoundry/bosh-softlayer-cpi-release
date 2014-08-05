package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/vm"
)

var _ = Describe("ConcreteFactoryOptions", func() {
	var (
		options ConcreteFactoryOptions

		validOptions = ConcreteFactoryOptions{
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
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			options = validOptions
		})

		It("does not return error if all fields are valid", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if StemcellsDir is empty", func() {
			options.StemcellsDir = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty StemcellsDir"))
		})

		It("returns error if DisksDir is empty", func() {
			options.DisksDir = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty DisksDir"))
		})

		It("returns error if HostEphemeralBindMountsDir is empty", func() {
			options.HostEphemeralBindMountsDir = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Must provide non-empty HostEphemeralBindMountsDir"))
		})

		It("returns error if HostPersistentBindMountsDir is empty", func() {
			options.HostPersistentBindMountsDir = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Must provide non-empty HostPersistentBindMountsDir"))
		})

		It("returns error if GuestEphemeralBindMountPath is empty", func() {
			options.GuestEphemeralBindMountPath = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Must provide non-empty GuestEphemeralBindMountPath"))
		})

		It("returns error if GuestPersistentBindMountsDir is empty", func() {
			options.GuestPersistentBindMountsDir = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Must provide non-empty GuestPersistentBindMountsDir"))
		})

		It("returns error if agent section is not valid", func() {
			options.Agent.Mbus = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Agent configuration"))
		})
	})
})
