package disk_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
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
		finder = NewFSFinder("/fake-disks-dir", fs, logger)
	})

	XDescribe("Find", func() {
		It("returns disk and found as true if disk path exists", func() {
			err := fs.WriteFile("/fake-disks-dir/fake-disk-id", []byte{})
			Expect(err).ToNot(HaveOccurred())

			disk, found, err := finder.Find(1234)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			expectedDisk := NewSoftLayerDisk(1234, logger)
			Expect(disk).To(Equal(expectedDisk))
		})

		It("returns found as false if disk path does not exist", func() {
			disk, found, err := finder.Find(1234)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
			Expect(disk).To(BeNil())
		})
	})
})
