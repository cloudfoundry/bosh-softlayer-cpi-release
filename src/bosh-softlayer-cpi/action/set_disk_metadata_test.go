package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/softlayer/disk_service"
	diskFakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	"fmt"
)

var _ = Describe("SetVMMetadata", func() {
	var (
		err          error
		diskCID      DiskCID
		diskMetadata DiskMetadata

		diskService     *diskFakes.FakeService
		setDiskMetadata SetDiskMetadata
	)

	BeforeEach(func() {
		diskCID = DiskCID(12345678)
		diskMetadata = DiskMetadata{
			"director":       "bats-director",
			"deployment":     "automation-cf",
			"instance_id":    "1f981b7e-1663-4aeb-8b7c-5e7d8783a4e3",
			"job":            "consul",
			"instance_index": "0",
			"instance_name":  "consul/1f981b7e-1663-4aeb-8b7c-5e7d8783a4e3",
			"attached_at":    "2017-09-12T14:43:41Z",
		}
		diskService = &diskFakes.FakeService{}
		setDiskMetadata = NewSetDiskMetadata(diskService)
	})

	Describe("Run", func() {
		It("set the vm metadata", func() {
			_, err = setDiskMetadata.Run(diskCID, diskMetadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(diskService.SetMetadataCallCount()).To(Equal(1))
			_, actualMetadata := diskService.SetMetadataArgsForCall(0)
			Expect(actualMetadata).To(Equal(disk.Metadata(diskMetadata)))
		})

		It("returns an error if vmService set metadata call returns an error", func() {
			diskService.SetMetadataReturns(
				errors.New("fake-vm-service-error"),
			)

			_, err = setDiskMetadata.Run(diskCID, diskMetadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(diskService.SetMetadataCallCount()).To(Equal(1))
		})

		It("returns an error if diskService set metadata call returns an api error", func() {
			diskService.SetMetadataReturns(
				api.NewDiskNotFoundError(diskCID.String(), false),
			)

			_, err = setDiskMetadata.Run(diskCID, diskMetadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Disk '%d' not found", diskCID)))
			Expect(diskService.SetMetadataCallCount()).To(Equal(1))
		})
	})
})
