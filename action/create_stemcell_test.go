package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"
	
	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
)

var _ = Describe("CreateStemcell", func() {
	var (
		stemcellImporter *fakestem.FakeImporter
		action           CreateStemcell
	)

	BeforeEach(func() {
		stemcellImporter = &fakestem.FakeImporter{}
		action = NewCreateStemcell(stemcellImporter)
	})

	Describe("Run", func() {
		It("returns id for created stemcell from image path", func() {
			stemcellImporter.ImportFromPathStemcell = fakestem.NewFakeStemcell("fake-stemcell-id")

			id, err := action.Run("/fake-image-path", CreateStemcellCloudProps{})
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(StemcellCID("fake-stemcell-id")))

			Expect(stemcellImporter.ImportFromPathImagePath).To(Equal("/fake-image-path"))
		})

		It("returns error if creating stemcell fails", func() {
			stemcellImporter.ImportFromPathErr = errors.New("fake-add-err")

			id, err := action.Run("/fake-image-path", CreateStemcellCloudProps{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-add-err"))
			Expect(id).To(Equal(StemcellCID("")))
		})
	})
})
