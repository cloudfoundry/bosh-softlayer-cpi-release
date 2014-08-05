package stemcell_test

import (
	"os"

	boshlog "bosh/logger"
	fakesys "bosh/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

var _ = Describe("FSFinder", func() {
	var (
		fs     *fakesys.FakeFileSystem
		logger boshlog.Logger
		finder FSFinder
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		finder = NewFSFinder("/fake-collection-dir", fs, logger)
	})

	Describe("Find", func() {
		It("returns stemcell and found as true if stemcell directory exists", func() {
			err := fs.MkdirAll("/fake-collection-dir/fake-stemcell-id", os.ModeDir)
			Expect(err).ToNot(HaveOccurred())

			stemcell, found, err := finder.Find("fake-stemcell-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			expectedStemcell := NewFSStemcell("fake-stemcell-id", "/fake-collection-dir/fake-stemcell-id", fs, logger)
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
