package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	"bosh-softlayer-cpi/registry"
	registryfakes "bosh-softlayer-cpi/registry/fakes"
	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
)

var _ = Describe("DetachDisk", func() {
	var (
		err                   error
		expectedAgentSettings registry.AgentSettings

		vmService      *instancefakes.FakeService
		registryClient *registryfakes.FakeClient

		detachDisk DetachDisk
	)
	BeforeEach(func() {
		vmService = &instancefakes.FakeService{}
		registryClient = &registryfakes.FakeClient{}
		detachDisk = NewDetachDisk(vmService, registryClient)
	})

	Describe("Run", func() {
		var (
			vmCID   VMCID
			diskCID DiskCID
		)

		BeforeEach(func() {
			vmCID = VMCID(12345678)
			diskCID = DiskCID(22345678)

			registryClient.FetchSettings = registry.AgentSettings{
				Disks: registry.DisksSettings{
					Persistent: map[string]registry.PersistentSettings{
						"22345678": {
							ID:       "22345678",
							VolumeID: "fake-device-name",
							Path:     "fake-volume-path",
						},
					},
				},
			}

			expectedAgentSettings = registry.AgentSettings{
				Disks: registry.DisksSettings{
					Persistent: map[string]registry.PersistentSettings{},
				},
			}
		})

		It("detaches the disk", func() {
			_, err = detachDisk.Run(vmCID, diskCID)
			Expect(err).NotTo(HaveOccurred())
			Expect(vmService.DetachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeTrue())
			Expect(registryClient.UpdateCalled).To(BeTrue())
			Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
		})

		It("returns an error if vmService detach disk call returns an error", func() {
			vmService.DetachDiskReturns(
				errors.New("fake-vm-service-error"),
			)

			_, err = detachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(vmService.DetachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeFalse())
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if registryClient fetch call returns an error", func() {
			registryClient.FetchErr = errors.New("fake-registry-client-error")

			_, err = detachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
			Expect(vmService.DetachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeTrue())
			Expect(registryClient.UpdateCalled).To(BeFalse())
		})

		It("returns an error if registryClient update call returns an error", func() {
			registryClient.UpdateErr = errors.New("fake-registry-client-error")

			_, err = detachDisk.Run(vmCID, diskCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
			Expect(vmService.DetachDiskCallCount()).To(Equal(1))
			Expect(registryClient.FetchCalled).To(BeTrue())
			Expect(registryClient.UpdateCalled).To(BeTrue())
		})

		Context("when Persistent contains two Settings", func() {
			BeforeEach(func() {
				registryClient.FetchSettings = registry.AgentSettings{
					Disks: registry.DisksSettings{
						Persistent: map[string]registry.PersistentSettings{
							"22345678": {
								ID:       "22345678",
								VolumeID: "fake-device-name1",
								Path:     "fake-volume-path1",
							},
							"32345678": {
								ID:       "32345678",
								VolumeID: "fake-device-name2",
								Path:     "fake-volume-path2",
							},
						},
					},
				}

				expectedAgentSettings = registry.AgentSettings{
					Disks: registry.DisksSettings{
						Persistent: map[string]registry.PersistentSettings{
							"32345678": {
								ID:       "32345678",
								VolumeID: "fake-device-name2",
								Path:     "fake-volume-path2",
							},
						},
					},
				}
			})

			It("detaches the disk and re-attaches left disk", func() {
				_, err = detachDisk.Run(vmCID, diskCID)
				Expect(err).NotTo(HaveOccurred())
				Expect(vmService.DetachDiskCallCount()).To(Equal(1))
				Expect(registryClient.FetchCalled).To(BeTrue())
				Expect(registryClient.UpdateCalled).To(BeTrue())
				Expect(vmService.ReAttachLeftDiskCallCount()).To(Equal(1))
				Expect(registryClient.UpdateSettings).To(Equal(expectedAgentSettings))
			})

			It("returns an error if vmService ReAttachLeftDisk call returns an error", func() {
				vmService.ReAttachLeftDiskReturns(
					errors.New("fake-vm-service-error"),
				)

				_, err = detachDisk.Run(vmCID, diskCID)
				Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
				Expect(vmService.DetachDiskCallCount()).To(Equal(1))
				Expect(registryClient.FetchCalled).To(BeTrue())
				Expect(registryClient.UpdateCalled).To(BeTrue())
				Expect(vmService.ReAttachLeftDiskCallCount()).To(Equal(1))
			})
		})
	})
})
