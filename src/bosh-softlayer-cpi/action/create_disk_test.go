package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	"bosh-softlayer-cpi/api"
	diskfakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("CreateDisk", func() {
	var (
		err     error
		diskCID string

		diskService *diskfakes.FakeService
		vmService   *instancefakes.FakeService
		createDisk  CreateDisk
	)
	BeforeEach(func() {
		diskService = &diskfakes.FakeService{}
		vmService = &instancefakes.FakeService{}
		createDisk = NewCreateDisk(diskService, vmService)
	})

	Describe("Run", func() {
		var (
			size       int
			cloudProps DiskCloudProperties
			vmCID      VMCID
		)
		BeforeEach(func() {
			size = 32768
			cloudProps = DiskCloudProperties{}
			vmCID = VMCID(0)
		})

		Context("when vmCID is set", func() {
			BeforeEach(func() {
				vmCID = VMCID(12345678)
				cloudProps = DiskCloudProperties{
					DataCenter: "fake-datacenter-name",
				}

				diskService.CreateReturns(
					22345678,
					nil,
				)
				vmService.FindReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(1234567),
						Datacenter: &datatypes.Location{
							Name: sl.String("fake-datacenter-name"),
						},
					},
					nil,
				)
			})

			It("creates the disk at the vm zone", func() {
				diskCID, err = createDisk.Run(size, cloudProps, vmCID)
				Expect(err).NotTo(HaveOccurred())
				Expect(vmService.FindCallCount()).To(Equal(1))
				actualCid := vmService.FindArgsForCall(0)
				Expect(actualCid).To(Equal(12345678))
				Expect(diskService.CreateCallCount()).To(Equal(1))
				actualSize, _, actualLocation, _ := diskService.CreateArgsForCall(0)
				Expect(actualSize).To(Equal(size))
				Expect(actualLocation).To(Equal("fake-datacenter-name"))
				Expect(diskCID).To(Equal("22345678"))
			})

			It("returns an error if vmService find call returns an error", func() {
				vmService.FindReturns(
					&datatypes.Virtual_Guest{},
					errors.New("fake-instance-service-error"),
				)

				_, err = createDisk.Run(size, cloudProps, vmCID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-instance-service-error"))
				Expect(vmService.FindCallCount()).To(Equal(1))
				Expect(diskService.CreateCallCount()).To(Equal(0))
			})

			It("returns an error if vmService find call returns an error", func() {
				vmService.FindReturns(
					&datatypes.Virtual_Guest{},
					api.NewDiskCreationFailedError("Not supported", false),
				)

				_, err = createDisk.Run(size, cloudProps, vmCID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Disk failed to create: "))
				Expect(vmService.FindCallCount()).To(Equal(1))
				Expect(diskService.CreateCallCount()).To(Equal(0))
			})

			It("returns an error if vmService create call returns an error", func() {
				vmService.FindReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(1234567),
						Datacenter: &datatypes.Location{
							Name: sl.String("fake-datacenter-name"),
						},
					},
					nil,
				)

				diskService.CreateReturns(
					0,
					errors.New("fake-instance-service-error"),
				)

				_, err = createDisk.Run(size, cloudProps, vmCID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-instance-service-error"))
				Expect(vmService.FindCallCount()).To(Equal(1))
				Expect(diskService.CreateCallCount()).To(Equal(1))
			})
		})
	})
})
