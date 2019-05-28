package action_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	"bosh-softlayer-cpi/api"
	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
	"fmt"
)

var _ = Describe("GetDisks", func() {
	var (
		err   error
		vmCID VMCID

		vmService *instancefakes.FakeService
		getDisks  GetDisks
	)

	BeforeEach(func() {
		vmCID = VMCID(12345678)

		vmService = &instancefakes.FakeService{}
		getDisks = NewGetDisks(vmService)
	})

	Describe("Run", func() {
		It("returns the list of attached disks", func() {
			vmService.AttachedDisksReturns(
				[]string{"fake-disk-1", "fake-disk-2"},
				nil,
			)
			_, err = getDisks.Run(vmCID)
			Expect(err).NotTo(HaveOccurred())
			Expect(vmService.AttachedDisksCallCount()).To(Equal(1))
		})

		It("returns an error if vmService attached disks call returns an error", func() {
			vmService.AttachedDisksReturns(
				[]string{},
				errors.New("fake-vm-service-error"),
			)

			_, err = getDisks.Run(vmCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(vmService.AttachedDisksCallCount()).To(Equal(1))
		})

		It("returns an error if vmService attached disks returns an api error", func() {
			vmService.AttachedDisksReturns(
				[]string{},
				api.NewVMNotFoundError(vmCID.String()),
			)

			_, err = getDisks.Run(vmCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("VM '%d' not found", vmCID)))
			Expect(vmService.AttachedDisksCallCount()).To(Equal(1))
		})
	})
})
