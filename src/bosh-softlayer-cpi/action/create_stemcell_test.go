package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
	"bosh-softlayer-cpi/api"
	stemcellfakes "bosh-softlayer-cpi/softlayer/stemcell_service/fakes"
)

var _ = Describe("CreateStemcell", func() {
	var (
		err               error
		createdStemcellId int
		stemcellCID       string
		cloudProps        StemcellCloudProperties

		stemcellService *stemcellfakes.FakeService
		createStemcell  CreateStemcellAction
	)

	BeforeEach(func() {
		stemcellService = &stemcellfakes.FakeService{}
		createStemcell = NewCreateStemcell(stemcellService)
	})

	Describe("Run", func() {
		Context("when infrastructure is not softlayer", func() {
			BeforeEach(func() {
				cloudProps.Infrastructure = "fake-insfrastructure"
			})

			It("returns an error", func() {
				_, err = createStemcell.Run("fake-stemcell-tarball", cloudProps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid 'fake-insfrastructure' infrastructure"))
				Expect(stemcellService.FindCallCount()).To(Equal(0))
				Expect(stemcellService.CreateFromTarballCallCount()).To(Equal(0))
			})
		})

		Context("from light-stemcell", func() {
			BeforeEach(func() {
				cloudProps = StemcellCloudProperties{
					Id:             12345678,
					Infrastructure: "softlayer",
					Uuid:           "fake-uuid",
					DatacenterName: "fake-datacenter-name",
				}
			})

			It("creates the stemcell", func() {
				stemcellService.FindReturns(
					"fake-global-identifier",
					nil,
				)

				stemcellCID, err = createStemcell.Run("fake-light-stemcell-imagePath", cloudProps)
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

			It("returns an error if stemcellService find returns an api error", func() {
				stemcellService.FindReturns(
					"",
					api.NewStemcellkNotFoundError("fake-stemcell-imagePath", false),
				)

				_, err = createStemcell.Run("fake-stemcell-imagePath", cloudProps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Stemcell 'fake-stemcell-imagePath' not found"))
				Expect(stemcellService.FindCallCount()).To(Equal(1))
			})
		})

		Context("from raw-stemcell", func() {
			BeforeEach(func() {
				cloudProps = StemcellCloudProperties{
					Infrastructure: "softlayer",
					DatacenterName: "fake-datacenter",
					OsCode:         "fake-os-code",
				}
				createdStemcellId = 12345678
			})

			It("creates the stemcell", func() {
				stemcellService.CreateFromTarballReturns(
					createdStemcellId,
					nil,
				)

				stemcellCID, err = createStemcell.Run("fake-light-stemcell-imagePath", cloudProps)
				Expect(err).NotTo(HaveOccurred())
				Expect(stemcellService.CreateFromTarballCallCount()).To(Equal(1))
				Expect(stemcellCID).To(Equal(StemcellCID(createdStemcellId).String()))
			})

			It("returns an error if stemcellService CreateFromTarball call returns an error", func() {
				stemcellService.CreateFromTarballReturns(
					0,
					errors.New("fake-stemcell-service-error"),
				)

				_, err = createStemcell.Run("fake-stemcell-imagePath", cloudProps)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stemcell-service-error"))
				Expect(stemcellService.CreateFromTarballCallCount()).To(Equal(1))
			})
		})
	})
})
