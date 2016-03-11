package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakestem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell/fakes"
)

var _ = Describe("DeleteStemcell", func() {
	var (
		stemcellFinder *fakestem.FakeFinder
		action         DeleteStemcell
	)

	BeforeEach(func() {
		stemcellFinder = &fakestem.FakeFinder{}
		action = NewDeleteStemcell(stemcellFinder)
	})

	Describe("Run", func() {
		It("tries to find stemcell with given stemcell cid", func() {
			_, err := action.Run(1234)
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcellFinder.FindID).To(Equal(1234))
		})

		Context("when stemcell is found with given stemcell cid", func() {
			var (
				stemcell *fakestem.FakeStemcell
			)

			BeforeEach(func() {
				stemcell = fakestem.NewFakeStemcell(1234, "fake-stemcell-id", fakestem.FakeStemcellKind)
				stemcellFinder.FindStemcell = stemcell
				stemcellFinder.FindFound = true
			})

			It("deletes stemcell", func() {
				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())

				Expect(stemcell.DeleteCalled).To(BeTrue())
			})

			It("returns error if deleting stemcell fails", func() {
				stemcell.DeleteErr = errors.New("fake-delete-err")

				_, err := action.Run(1234)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-err"))
			})
		})

		Context("when stemcell is not found with given cid", func() {
			It("does not return error", func() {
				stemcellFinder.FindFound = false

				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when stemcell finding fails", func() {
			It("does not return error", func() {
				stemcellFinder.FindErr = errors.New("fake-find-err")

				_, err := action.Run(1234)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
			})
		})
	})
})
