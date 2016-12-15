package action_test

import (
        "errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "bosh-softlayer-cpi/action"

        fakeDisk "bosh-softlayer-cpi/softlayer/disk/fakes"
)

var _ = Describe("HasDisk", func() {
	var (
		fakeDiskFinder *fakeDisk.FakeDiskFinder
		action       HasDiskAction

	)

	BeforeEach(func() {
		fakeDiskFinder = &fakeDisk.FakeDiskFinder{}
		action = NewHasDisk(fakeDiskFinder)
	})

	Describe("Run", func() {
		var (
			diskCid DiskCID
			found bool
			err   error
		)

		BeforeEach(func() {
			diskCid = DiskCID(123456)
		})

		JustBeforeEach(func() {
			found, err = action.Run(diskCid)
		})
		
		Context("when has disk succeeds", func() {
			BeforeEach(func() {
				fakeDiskFinder.FindReturns(nil, true, nil)
			})

			It("returns no error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when no error occurs but diskID is 0", func() {
			BeforeEach(func() {
				fakeDiskFinder.FindReturns(nil, false, nil)
			})

			It("returns no error but still not found", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when has disk fails", func() {
			BeforeEach(func() {
				fakeDiskFinder.FindReturns(nil, false, errors.New("disk not found"))
			})

			It("returns no error", func() {
				Expect(err).To(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})

