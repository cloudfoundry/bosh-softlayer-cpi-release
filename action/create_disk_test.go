package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
	fakedisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk/fakes"

	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
)

var _ = Describe("CreateDisk", func() {
	var (
		fakeVmFinder    *fakescommon.FakeVMFinder
		fakeVm          *fakescommon.FakeVM
		fakeDisk        *fakedisk.FakeDisk
		action          CreateDiskAction
		fakeDiskCreator *fakedisk.FakeDiskCreator

		diskCloudProp bslcdisk.DiskCloudProperties
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVm = &fakescommon.FakeVM{}
		fakeDiskCreator = &fakedisk.FakeDiskCreator{}
		fakeDisk = &fakedisk.FakeDisk{}
		action = NewCreateDisk(fakeVmFinder, fakeDiskCreator)
		diskCloudProp = bslcdisk.DiskCloudProperties{}
	})

	Describe("Run", func() {
		var (
			diskCidStr string
			err        error
			vmCid      VMCID
		)

		BeforeEach(func() {
			vmCid = VMCID(123456)
		})

		JustBeforeEach(func() {
			diskCidStr, err = action.Run(100, diskCloudProp, vmCid)
		})

		Context("when create disk succeeds", func() {
			BeforeEach(func() {
				fakeVm.GetDataCenterIdReturns(123456)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeDiskCreator.CreateReturns(fakeDisk, nil)
			})

			It("fetches vm by cid", func() {
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				actualCid := fakeVmFinder.FindArgsForCall(0)
				Expect(actualCid).To(Equal(123456))
			})

			It("no error return", func() {
				Expect(fakeDiskCreator.CreateCallCount()).To(Equal(1))
				actualDiskSize, actualDiskCloudProperties, actualDataCenterId := fakeDiskCreator.CreateArgsForCall(0)
				Expect(actualDiskSize).To(Equal(100))
				Expect(actualDiskCloudProperties).To(Equal(diskCloudProp))
				Expect(actualDataCenterId).To(Equal(123456))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when find vm error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})

		Context("when find vm return false", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, nil)
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when create disk error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeDiskCreator.CreateReturns(nil, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
