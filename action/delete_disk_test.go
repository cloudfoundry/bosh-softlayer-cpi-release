package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakedisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk/fakes"
)

var _ = Describe("DeleteDisk", func() {
	var (
		fakeDiskFinder *fakedisk.FakeDiskFinder
		fakeDisk       *fakedisk.FakeDisk
		action         DeleteDiskAction
	)

	BeforeEach(func() {
		fakeDiskFinder = &fakedisk.FakeDiskFinder{}
		fakeDisk = &fakedisk.FakeDisk{}
		action = NewDeleteDisk(fakeDiskFinder)
	})

	Describe("Run", func() {
		var (
			diskCid DiskCID
			err     error
		)

		BeforeEach(func() {
			diskCid = DiskCID(123456)
		})

		JustBeforeEach(func() {
			_, err = action.Run(diskCid)
		})

		Context("when delete disk succeeds", func() {
			BeforeEach(func() {
				fakeDisk.DeleteReturns(nil)
				fakeDiskFinder.FindReturns(fakeDisk, true, nil)
			})

			It("find disk by diskCid", func() {
				Expect(fakeDiskFinder.FindCallCount()).To(Equal(1))
				actualDiskCid := fakeDiskFinder.FindArgsForCall(0)
				Expect(actualDiskCid).To(Equal(int(diskCid)))
			})

			It("no err return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when disk is not found", func() {
			BeforeEach(func() {
				fakeDiskFinder.FindReturns(nil, false, nil)
			})

			It("no err return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when find disk error", func() {
			BeforeEach(func() {
				fakeDiskFinder.FindReturns(nil, false, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})

		Context("when delete disk error out", func() {
			BeforeEach(func() {
				fakeDisk.DeleteReturns(errors.New("kaboom"))
				fakeDiskFinder.FindReturns(fakeDisk, true, nil)
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
})
