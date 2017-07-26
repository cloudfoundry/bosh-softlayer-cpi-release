package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	stemcellfakes "bosh-softlayer-cpi/softlayer/stemcell_service/fakes"
)

var _ = Describe("CreateStemcell", func() {
	var (
		err         error
		stemcellCID string
		cloudProps  CreateStemcellCloudProps

		stemcellService *stemcellfakes.FakeService
		createStemcell  CreateStemcellAction
	)

	BeforeEach(func() {
		stemcellService = &stemcellfakes.FakeService{}
		createStemcell = NewCreateStemcell(stemcellService)
	})

	Describe("Run", func() {
		BeforeEach(func() {
			cloudProps = CreateStemcellCloudProps{
				Id:             12345678,
				Uuid:           "fake-uuid",
				DatacenterName: "fake-datacenter-name",
			}
		})

		It("creates the stemcell", func() {
			stemcellService.FindReturns(
				"fake-global-identifier",
				nil,
			)

			stemcellCID, err = createStemcell.Run("fake-stemcell-imagePath", cloudProps)
			Expect(err).NotTo(HaveOccurred())
			Expect(stemcellService.FindCallCount()).To(Equal(1))
			Expect(stemcellCID).To(Equal(StemcellCID(cloudProps.Id).String()))
		})

		It("returns an error if stemcellService find call returns an error", func() {
			stemcellService.FindReturns(
				"",
				errors.New("fake-stemcell-service-error"),
			)

			stemcellCID, err = createStemcell.Run("fake-stemcell-imagePath", cloudProps)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-stemcell-service-error"))
			Expect(stemcellService.FindCallCount()).To(Equal(1))
			Expect(stemcellCID).NotTo(Equal(StemcellCID(cloudProps.Id).String()))
		})
	})
})
