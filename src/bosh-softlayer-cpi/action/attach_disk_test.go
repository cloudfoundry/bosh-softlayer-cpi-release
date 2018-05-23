package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"
	registryfakes "bosh-softlayer-cpi/registry/fakes"
	diskfakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
	"fmt"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("AttachDisk", func() {
	var (
		err                   error
		expectedAgentSettings registry.AgentSettings

		diskService    *diskfakes.FakeService
		vmService      *instancefakes.FakeService
		registryClient *registryfakes.FakeClient
		attachDisk     AttachDisk
	)
	BeforeEach(func() {
		diskService = &diskfakes.FakeService{}
		vmService = &instancefakes.FakeService{}
		registryClient = &registryfakes.FakeClient{}
		attachDisk = NewAttachDisk(diskService, vmService, registryClient)
	})

	Describe("Run", func() {
		var (
			vmCID   VMCID
			diskCID DiskCID
		)

		BeforeEach(func() {
			vmCID = VMCID(12345678)
			diskCID = DiskCID(25667635)

			expectedAgentSettings = registry.AgentSettings{
				Disks: registry.DisksSettings{
					Persistent: map[string]registry.PersistentSettings{
						"25667635": {
							ID: "25667635",
							ISCSISettings: registry.ISCSISettings{
								InitiatorName: "iqn.yyyy-mm.fake-domain:fake-username",
								Target:        "10.1.22.170",
								Username:      "fake-username",
								Password:      "fake-password",
							},
						},
					},
				},
			}

			diskService.FindReturns(
				&datatypes.Network_Storage{
					Id: sl.Int(1234567),
				},
				nil,
			)
			vmService.AttachDiskReturns(
				[]byte(`{"id":"25667635","iscsi_settings":{"initiator_name":"iqn.yyyy-mm.fake-domain:fake-username","target":"10.1.22.170","username":"fake-username","password":"fake-password"}}`),
				nil,
			)
		})

		It("attaches the disk", func() {
			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).NotTo(HaveOccurred())
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(vmService.AttachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeTrue())
			Expect(registryClient.UpdateCalled).To(BeTrue())
			Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
		})

		It("returns an error if diskService find call returns an error", func() {
			diskService.FindReturns(
				&datatypes.Network_Storage{},
				errors.New("fake-disk-service-error"),
			)

			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-disk-service-error"))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(vmService.AttachDiskCallCount()).To(Equal(0))
			Expect(registryClient.FetchCalled).To(BeFalse())
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if diskService find returns an api error", func() {
			diskService.FindReturns(
				&datatypes.Network_Storage{},
				api.NewVMNotFoundError(vmCID.String()),
			)

			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("VM '%d' not found", vmCID)))
			Expect(diskService.FindCallCount()).To(Equal(1))
		})

		It("returns an error if vmService attach disk call returns an error", func() {
			vmService.AttachDiskReturns(
				[]byte{},
				errors.New("fake-vm-service-error"),
			)

			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(vmService.AttachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeFalse())
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if vmService attach disk returns an api error", func() {
			vmService.AttachDiskReturns(
				[]byte{},
				api.NewVMNotFoundError(vmCID.String()),
			)

			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("VM '%d' not found", vmCID)))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(vmService.AttachDiskCallCount()).To(Equal(1))
		})

		It("returns an error if registryClient fetch call returns an error", func() {
			registryClient.FetchErr = errors.New("fake-registry-client-error")

			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(vmService.AttachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeTrue())
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if registryClient update call returns an error", func() {
			registryClient.UpdateErr = errors.New("fake-registry-client-error")

			_, err = attachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(vmService.AttachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeTrue())
			Expect(registryClient.UpdateCalled).To(BeTrue())
		})
	})
})
