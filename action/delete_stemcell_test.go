package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
)

var _ = Describe("DeleteStemcell", func() {
	var (
		stemcellFinder *fakestem.FakeFinder
		action         DeleteStemcell
		logger         boshlog.Logger
	)

	BeforeEach(func() {
		stemcellFinder = &fakestem.FakeFinder{}

		logger = boshlog.NewLogger(boshlog.LevelNone)
		action = NewDeleteStemcell(stemcellFinder, logger)
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

			It("does not delete stemcell", func() {
				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())

				Expect(stemcell.DeleteCalled).To(BeFalse())
			})

			It("logs instead of returning error if deleting stemcell fails", func() {
				stemcell.DeleteErr = errors.New("fake-delete-err")

				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
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
			It("logs instead of returning error", func() {
				stemcellFinder.FindErr = errors.New("fake-find-err")

				_, err := action.Run(1234)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
