package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"

	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

var _ = Describe("SetVMMetadata", func() {
	var (
		err        error
		vmCID      VMCID
		vmMetadata VMMetadata

		vmService *instancefakes.FakeService

		setVMMetadata SetVMMetadata
	)

	BeforeEach(func() {
		vmCID = VMCID(12345678)
		vmMetadata = VMMetadata{
			"deployment": "fake-deployment",
			"job":        "fake-job",
			"index":      "fake-index",
		}
		vmService = &instancefakes.FakeService{}
		setVMMetadata = NewSetVMMetadata(vmService)
	})

	Describe("Run", func() {
		It("set the vm metadata", func() {
			_, err = setVMMetadata.Run(vmCID, vmMetadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(vmService.SetMetadataCallCount()).To(Equal(1))
			_, actualMetadata := vmService.SetMetadataArgsForCall(0)
			Expect(actualMetadata).To(Equal(instance.Metadata(vmMetadata)))
		})

		It("returns an error if vmService set metadata call returns an error", func() {
			vmService.SetMetadataReturns(
				errors.New("fake-vm-service-error"),
			)

			_, err = setVMMetadata.Run(vmCID, vmMetadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-vm-service-error"))
			Expect(vmService.SetMetadataCallCount()).To(Equal(1))
		})
	})
})
