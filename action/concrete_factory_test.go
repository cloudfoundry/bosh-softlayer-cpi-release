package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var _ = Describe("ConcreteFactory", func() {
	var (
		logger boshlog.Logger

		options = ConcreteFactoryOptions{
			StemcellsDir: "/tmp/stemcells",
		}

		factory Factory
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		factory = NewConcreteFactory(
			options,
			logger,
		)
	})

	Context("Stemcell methods", func() {
		It("create_stemcell", func() {
			action, err := factory.Create("create_stemcell")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("delete_stemcell", func() {
			action, err := factory.Create("delete_stemcell")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("VM methods", func() {
		It("create_vm", func() {
			action, err := factory.Create("create_vm")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("delete_vm", func() {
			action, err := factory.Create("delete_vm")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("has_vm", func() {
			action, err := factory.Create("has_vm")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("reboot_vm", func() {
			action, err := factory.Create("reboot_vm")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("set_vm_metadata", func() {
			action, err := factory.Create("set_vm_metadata")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("configure_networks", func() {
			action, err := factory.Create("configure_networks")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Disk methods", func() {
		It("creates an iSCSI disk", func() {
			action, err := factory.Create("create_disk")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("deletes the detached iSCSI disk", func() {
			action, err := factory.Create("delete_disk")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("attaches an iSCSI disk to a virtual guest", func() {
			action, err := factory.Create("attach_disk")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("detaches the iSCSI disk from virtual guest", func() {
			action, err := factory.Create("detach_disk")
			Expect(action).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Unsupported methods", func() {
		It("returns error because CPI machine is not self-aware if action is current_vm_id", func() {
			action, err := factory.Create("current_vm_id")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error because snapshotting is not implemented if action is snapshot_disk", func() {
			action, err := factory.Create("snapshot_disk")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error because snapshotting is not implemented if action is delete_snapshot", func() {
			action, err := factory.Create("delete_snapshot")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error since CPI should not keep state if action is get_disks", func() {
			action, err := factory.Create("get_disks")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error because ping is not official CPI method if action is ping", func() {
			action, err := factory.Create("ping")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})
	})

	Context("Misc", func() {
		It("returns error if action cannot be created", func() {
			action, err := factory.Create("fake-unknown-action")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})
	})
})
