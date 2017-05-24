package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("HasVM", func() {
	var (
		err   error
		found bool

		vmService *instancefakes.FakeService

		hasVM HasVM
	)

	BeforeEach(func() {
		vmService = &instancefakes.FakeService{}
		hasVM = NewHasVM(vmService)
	})

	Describe("Run", func() {
		It("returns true if vm ID exist", func() {
			vmService.FindReturns(datatypes.Virtual_Guest{
				Id: sl.Int(1234567),
			},
				true,
				nil)

			found, err = hasVM.Run(1234567)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())

			Expect(vmService.FindCallCount()).To(Equal(1))
			Expect(vmService.FindArgsForCall(0)).To(Equal(1234567))
		})

		It("returns false if vm ID does not exist", func() {
			vmService.FindReturns(datatypes.Virtual_Guest{},
				false,
				nil)

			found, err = hasVM.Run(1234567)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())

			Expect(vmService.FindCallCount()).To(Equal(1))
			Expect(vmService.FindArgsForCall(0)).To(Equal(1234567))
		})

		It("returns an error if vmService find call returns an error", func() {
			vmService.FindReturns(datatypes.Virtual_Guest{},
				false,
				errors.New("fake-vm-service-error"))

			_, err = hasVM.Run(1234567)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))

			Expect(vmService.FindCallCount()).To(Equal(1))
			Expect(vmService.FindArgsForCall(0)).To(Equal(1234567))
		})
	})
})
