package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"
	fakedisk "github.com/maximilien/bosh-softlayer-cpi/disk/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/vm/fakes"
)

var _ = Describe("AttachDisk", func() {
	var (
		vmFinder   *fakevm.FakeFinder
		diskFinder *fakedisk.FakeFinder
		action     AttachDisk
	)

	BeforeEach(func() {
		vmFinder = &fakevm.FakeFinder{}
		diskFinder = &fakedisk.FakeFinder{}
		action = NewAttachDisk(vmFinder, diskFinder)
	})

	Describe("Run", func() {
		It("tries to find VM with given VM cid", func() {
			vmFinder.FindFound = true
			vmFinder.FindVM = fakevm.NewFakeVM("fake-vm-id")

			diskFinder.FindFound = true
			diskFinder.FindDisk = fakedisk.NewFakeDisk("fake-disk-id")

			_, err := action.Run("fake-vm-id", "fake-disk-id")
			Expect(err).ToNot(HaveOccurred())

			Expect(vmFinder.FindID).To(Equal("fake-vm-id"))
		})

		Context("when VM is found with given VM cid", func() {
			var (
				vm *fakevm.FakeVM
			)

			BeforeEach(func() {
				vm = fakevm.NewFakeVM("fake-vm-id")
				vmFinder.FindVM = vm
				vmFinder.FindFound = true
			})

			It("tries to find disk with given disk cid", func() {
				diskFinder.FindFound = true

				_, err := action.Run("fake-vm-id", "fake-disk-id")
				Expect(err).ToNot(HaveOccurred())

				Expect(diskFinder.FindID).To(Equal("fake-disk-id"))
			})

			Context("when disk is found with given disk cid", func() {
				var (
					disk *fakedisk.FakeDisk
				)

				BeforeEach(func() {
					disk = fakedisk.NewFakeDisk("fake-disk-id")
					diskFinder.FindDisk = disk
					diskFinder.FindFound = true
				})

				It("does not return error when attaching found disk to found VM succeeds", func() {
					_, err := action.Run("fake-vm-id", "fake-disk-id")
					Expect(err).ToNot(HaveOccurred())

					Expect(vm.AttachDiskDisk).To(Equal(disk))
				})

				It("returns error if attaching disk fails", func() {
					vm.AttachDiskErr = errors.New("fake-attach-disk-err")

					_, err := action.Run("fake-vm-id", "fake-disk-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-attach-disk-err"))
				})
			})

			Context("when disk is not found with given cid", func() {
				It("returns error", func() {
					diskFinder.FindFound = false

					_, err := action.Run("fake-vm-id", "fake-disk-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Expected to find disk"))
				})
			})

			Context("when disk finding fails", func() {
				It("returns error", func() {
					diskFinder.FindErr = errors.New("fake-find-err")

					_, err := action.Run("fake-vm-id", "fake-disk-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-find-err"))
				})
			})
		})

		Context("when VM is not found with given cid", func() {
			It("returns error because disk can only be attached to an existing VM", func() {
				vmFinder.FindFound = false

				_, err := action.Run("fake-vm-id", "fake-disk-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Expected to find VM"))
			})
		})

		Context("when VM finding fails", func() {
			It("returns error because disk can only be attached to an existing VM", func() {
				vmFinder.FindErr = errors.New("fake-find-err")

				_, err := action.Run("fake-vm-id", "fake-disk-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-find-err"))
			})
		})
	})
})
