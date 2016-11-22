package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakestem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell/fakes"
)

var _ = Describe("CreateStemcell", func() {
	var (
		stemcellFinder *fakestem.FakeStemcellFinder
		action         CreateStemcellAction
	)

	BeforeEach(func() {
		stemcellFinder = &fakestem.FakeStemcellFinder{}
		action = NewCreateStemcell(stemcellFinder)
	})

	Describe("Run", func() {
		It("returns id for created stemcell from image path", func() {
			stemcellFinder.FindErr = nil
			stemcellFinder.FindStemcell = fakestem.NewFakeStemcell(1234, "fake-stemcell-id")

			id, err := action.Run("fake-path", CreateStemcellCloudProps{Uuid: "fake-stemcell-id"})
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(StemcellCID(1234).String()))
		})

		It("returns error if creating stemcell fails", func() {
			stemcellFinder.FindErr = errors.New("fake-add-err")

			id, err := action.Run("fake-path", CreateStemcellCloudProps{Uuid: "fake-stemcell-id"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-add-err"))
			Expect(id).To(Equal(StemcellCID(0).String()))
		})
	})
})
