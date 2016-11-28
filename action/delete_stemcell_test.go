package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakestem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell/fakes"
)

var _ = Describe("DeleteStemcell", func() {
	var (
		stemcellFinder *fakestem.FakeStemcellFinder
		action         DeleteStemcellAction
		logger         boshlog.Logger
	)

	BeforeEach(func() {
		stemcellFinder = &fakestem.FakeStemcellFinder{}

		logger = boshlog.NewLogger(boshlog.LevelNone)
		action = NewDeleteStemcell(stemcellFinder, logger)
	})

	Describe("Run", func() {
		var (
			stemcellCid StemcellCID
			err         error
		)

		BeforeEach(func() {
			stemcellCid = StemcellCID(1234567)
		})

		JustBeforeEach(func() {
			_, err = action.Run(stemcellCid)
		})

		Context("when delete stemcell always succeeds", func() {
			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
