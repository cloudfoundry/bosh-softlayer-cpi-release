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
		fakeStemcellFinder *fakestem.FakeStemcellFinder
		fakeStemcell       *fakestem.FakeStemcell

		action CreateStemcellAction
	)

	BeforeEach(func() {
		fakeStemcellFinder = &fakestem.FakeStemcellFinder{}
		fakeStemcell = &fakestem.FakeStemcell{}
		action = NewCreateStemcell(fakeStemcellFinder)
	})

	Describe("Run", func() {
		var (
			stemcellIdStr string
			err           error
		)

		JustBeforeEach(func() {
			stemcellIdStr, err = action.Run("fake-path", CreateStemcellCloudProps{Uuid: "fake-stemcell-id", Id: 123456})
		})

		Context("when create stemcell succeeds", func() {
			BeforeEach(func() {
				fakeStemcellFinder.FindByIdReturns(fakeStemcell, nil)
			})

			It("find stemcell by id", func() {
				Expect(fakeStemcellFinder.FindByIdCallCount()).To(Equal(1))
				acutalStemcellId := fakeStemcellFinder.FindByIdArgsForCall(0)
				Expect(acutalStemcellId).To(Equal(123456))
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when find stemcell error return", func() {
			BeforeEach(func() {
				fakeStemcellFinder.FindByIdReturns(nil, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
})
