package disk_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

var _ = XDescribe("FSDisk", func() {
	var (
		fs   *fakesys.FakeFileSystem
		disk FSDisk
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		disk = NewFSDisk(1234, "/fake-disk-path", fs, logger)
	})

	Describe("Delete", func() {
		It("deletes path", func() {
			err := fs.WriteFileString("/fake-disk-path", "fake-content")
			Expect(err).ToNot(HaveOccurred())

			err = disk.Delete()
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/fake-disk-path")).To(BeFalse())
		})

		It("returns error if deleting path fails", func() {
			fs.RemoveAllError = errors.New("fake-remove-all-err")

			err := disk.Delete()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-remove-all-err"))
		})
	})
})
