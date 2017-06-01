package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"

	registryfakes "bosh-softlayer-cpi/registry/fakes"
)

var _ = Describe("DeleteVM", func() {
	var (
		err   error
		vmCID VMCID

		vmService      *instancefakes.FakeService
		registryClient *registryfakes.FakeClient

		deleteVM DeleteVMAction
	)

	BeforeEach(func() {
		vmCID = VMCID(12345678)
		vmService = &instancefakes.FakeService{}
		registryClient = &registryfakes.FakeClient{}
		deleteVM = NewDeleteVM(vmService, registryClient)
	})

	Describe("Run", func() {
		It("deletes the vm", func() {
			_, err = deleteVM.Run(vmCID)
			Expect(err).NotTo(HaveOccurred())
			Expect(vmService.DeleteCallCount()).To(Equal(1))
			Expect(registryClient.DeleteCalled).To(BeTrue())
		})

		It("returns an error if vmService delete call returns an error", func() {
			vmService.DeleteReturns(
				errors.New("fake-vm-service-error"),
			)

			_, err = deleteVM.Run(vmCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(vmService.DeleteCallCount()).To(Equal(1))
			Expect(registryClient.DeleteCalled).To(BeFalse())
		})

		It("returns an error if registryClient delete call returns an error", func() {
			registryClient.DeleteErr = errors.New("fake-registry-client-error")

			_, err = deleteVM.Run(vmCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-client-error"))
			Expect(vmService.DeleteCallCount()).To(Equal(1))
			Expect(registryClient.DeleteCalled).To(BeTrue())
		})
	})
})
