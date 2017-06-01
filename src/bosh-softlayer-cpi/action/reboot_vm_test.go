package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
)

var _ = Describe("RebootVM", func() {
	var (
		err error
		vmCID VMCID
		vmService *instancefakes.FakeService

		rebootVM RebootVM
	)

	BeforeEach(func() {
		vmCID = VMCID(12345678)
		vmService = &instancefakes.FakeService{}
		rebootVM = NewRebootVM(vmService)
	})

	Describe("Run", func() {
		It("reboots the vm", func() {
			_, err = rebootVM.Run(vmCID)
			Expect(err).NotTo(HaveOccurred())
			Expect(vmService.RebootCallCount()).To(Equal(1))
		})

		It("returns an error if vmService reboot call returns an error", func() {
			vmService.RebootReturns(
				errors.New("fake-vm-service-error"),
			)

			_, err = rebootVM.Run(vmCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(vmService.RebootCallCount()).To(Equal(1))
		})
	})
})
