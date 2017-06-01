package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	imagefakes "bosh-softlayer-cpi/softlayer/stemcell_service/fakes"
)

var _ = Describe("DeleteStemcell", func() {
	var (
		err        error
		stemcellID StemcellCID

		imageService *imagefakes.FakeService

		deleteStemcell DeleteStemcellAction
	)

	BeforeEach(func() {
		stemcellID = StemcellCID(12345678)
		imageService = &imagefakes.FakeService{}
		deleteStemcell = NewDeleteStemcell(imageService)
	})

	Describe("Run", func() {
		It("deletes the stemcell", func() {
			_, err = deleteStemcell.Run(stemcellID)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
