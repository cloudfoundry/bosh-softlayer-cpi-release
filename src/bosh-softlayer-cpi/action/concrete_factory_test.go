package action_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	"bosh-softlayer-cpi/config"
	cpiLog "bosh-softlayer-cpi/logger"
	"bosh-softlayer-cpi/registry"
	bosl "bosh-softlayer-cpi/softlayer/client"
	boslconfig "bosh-softlayer-cpi/softlayer/config"
	"bosh-softlayer-cpi/softlayer/disk_service"
	"bosh-softlayer-cpi/softlayer/stemcell_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

var _ = Describe("ConcreteFactory", func() {
	var (
		uuidGen         *fakeuuid.FakeGenerator
		softlayerClient bosl.Client
		logger          cpiLog.Logger

		cfg = config.Config{}

		factory Factory
	)

	var (
		diskService      disk.Service
		imageService     stemcell.Service
		registryClient   registry.Client
		registryOptions  registry.ClientOptions
		agentOptions     registry.AgentOptions
		softlayerOptions boslconfig.Config
		vmService        instance.Service
	)

	BeforeEach(func() {
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = cpiLog.NewLogger(boshlog.LevelNone, "")
		cfg = config.Config{
			Cloud: config.Cloud{
				Properties: config.CPIProperties{
					Registry: registry.ClientOptions{
						Protocol: "http",
						Address:  "fake-host",
						Port:     5555,
						Username: "fake-username",
						Password: "fake-password",
					},
				},
			},
		}
		registryOptions = registry.ClientOptions{
			Protocol: "http",
			Address:  "fake-host",
			Port:     5555,
			Username: "fake-username",
			Password: "fake-password",
		}

		factory = NewConcreteFactory(
			softlayerClient,
			uuidGen,
			cfg,
			logger,
		)
	})

	BeforeEach(func() {
		diskService = disk.NewSoftlayerDiskService(
			softlayerClient,
			logger,
		)

		imageService = stemcell.NewSoftlayerStemcellService(
			softlayerClient,
			uuidGen,
			logger,
		)

		registryClient = registry.NewHTTPClient(
			cfg.Cloud.Properties.Registry,
			logger,
		)

		vmService = instance.NewSoftLayerVirtualGuestService(
			softlayerClient,
			uuidGen,
			logger,
		)
	})

	It("returns error if action cannot be created", func() {
		action, err := factory.Create("fake-unknown-action")
		Expect(err).To(HaveOccurred())
		Expect(action).To(BeNil())
	})

	It("create_disk", func() {
		action, err := factory.Create("create_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewCreateDisk(
			diskService,
			vmService,
		)))
	})

	It("delete_disk", func() {
		action, err := factory.Create("delete_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewDeleteDisk(diskService)))
	})

	It("attach_disk", func() {
		action, err := factory.Create("attach_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewAttachDisk(diskService, vmService, registryClient)))
	})

	It("detach_disk", func() {
		action, err := factory.Create("detach_disk")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewDetachDisk(vmService, registryClient)))
	})

	It("create_stemcell", func() {
		action, err := factory.Create("create_stemcell")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewCreateStemcell(imageService)))
	})

	It("delete_stemcell", func() {
		action, err := factory.Create("delete_stemcell")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewDeleteStemcell(imageService)))
	})

	It("create_vm", func() {
		action, err := factory.Create("create_vm")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewCreateVM(
			imageService,
			vmService,
			registryClient,
			registryOptions,
			agentOptions,
			softlayerOptions,
		)))
	})

	It("configure_networks", func() {
		action, err := factory.Create("configure_networks")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewConfigureNetworks(vmService, registryClient)))
	})

	It("delete_vm", func() {
		action, err := factory.Create("delete_vm")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewDeleteVM(vmService, registryClient, softlayerOptions)))
	})

	It("reboot_vm", func() {
		action, err := factory.Create("reboot_vm")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewRebootVM(vmService)))
	})

	It("set_vm_metadata", func() {
		action, err := factory.Create("set_vm_metadata")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewSetVMMetadata(vmService)))
	})

	It("has_vm", func() {
		action, err := factory.Create("has_vm")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewHasVM(vmService)))
	})

	It("get_disks", func() {
		action, err := factory.Create("get_disks")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewGetDisks(vmService)))
	})

	It("set_disk_metadata", func() {
		action, err := factory.Create("set_disk_metadata")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewSetDiskMetadata(diskService)))
	})

	It("info", func() {
		action, err := factory.Create("info")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewInfo()))
	})

	It("ping", func() {
		action, err := factory.Create("ping")
		Expect(err).ToNot(HaveOccurred())
		Expect(action).To(Equal(NewPing()))
	})

	It("when action is current_vm_id returns an error because this CPI does not implement the method", func() {
		action, err := factory.Create("current_vm_id")
		Expect(err).To(HaveOccurred())
		Expect(action).To(BeNil())
	})

	It("when action is wrong returns an error because it is not an official CPI method", func() {
		action, err := factory.Create("wrong")
		Expect(err).To(HaveOccurred())
		Expect(action).To(BeNil())
	})
})
