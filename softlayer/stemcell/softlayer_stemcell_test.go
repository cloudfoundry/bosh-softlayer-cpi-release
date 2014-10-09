package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

var _ = XDescribe("SoftLayerImporter", func() {
	var (
		stemcell SoftLayerStemcell
		logger   boshlog.Logger
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		stemcell = NewSoftLayerStemcell(1234, "/fake-stemcell-dir", logger)
	})

	Describe("Delete", func() {
		It("deletes directory in collection directory that contains unpacked stemcell", func() {
			err := stemcell.Delete()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if deleting stemcell directory fails", func() {
			err := stemcell.Delete()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-remove-all-err"))
		})
	})
})
