package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	slfakes "github.com/maximilien/softlayer-go/client/fakes"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
)

var _ = XDescribe("FSFinder", func() {
	var (
		softLayerClient softlayer.Client
		logger          boshlog.Logger
		finder          FSFinder
	)

	BeforeEach(func() {
		softLayerClient = slfakes.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		finder = NewFSFinder(softLayerClient, logger)
	})

	Describe("Find", func() {
		It("returns stemcell and found as true if stemcell directory exists", func() {
			stemcell, found, err := finder.Find("fake-stemcell-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			expectedStemcell := NewFSStemcell("fake-stemcell-id", logger)
			Expect(stemcell).To(Equal(expectedStemcell))
		})

		It("returns found as false if stemcell directory does not exist", func() {
			stemcell, found, err := finder.Find("fake-stemcell-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
			Expect(stemcell).To(BeNil())
		})
	})
})
