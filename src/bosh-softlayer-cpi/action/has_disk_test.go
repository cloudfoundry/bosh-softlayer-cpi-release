package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	diskfakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("HasDisk", func() {
	var (
		err   error
		found bool

		diskService *diskfakes.FakeService

		hasDisk HasDisk
	)

	BeforeEach(func() {
		diskService = &diskfakes.FakeService{}
		hasDisk = NewHasDisk(diskService)
	})

	Describe("Run", func() {
		It("returns true if disk ID exist", func() {
			diskService.FindReturns(
				datatypes.Network_Storage{
					Id: sl.Int(1234567),
				},
				true,
				nil,
			)

			found, err = hasDisk.Run(1234567)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())

			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(diskService.FindArgsForCall(0)).To(Equal(1234567))

		})

		It("returns false if disk ID does not exist", func() {
			diskService.FindReturns(
				datatypes.Network_Storage{},
				false,
				nil,
			)

			found, err = hasDisk.Run(1234567)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())

			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(diskService.FindArgsForCall(0)).To(Equal(1234567))
		})

		It("returns an error if diskService find call returns an error", func() {
			diskService.FindReturns(
				datatypes.Network_Storage{},
				false,
				errors.New("fake-vm-service-error"),
			)

			_, err = hasDisk.Run(1234567)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))

			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(diskService.FindArgsForCall(0)).To(Equal(1234567))
		})
	})
})
