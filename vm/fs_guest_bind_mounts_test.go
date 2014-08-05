package vm_test

import (
	boshlog "bosh/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/vm"
)

var _ = Describe("FSGuestBindMounts", func() {
	var (
		guestBindMounts FSGuestBindMounts
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		guestBindMounts = NewFSGuestBindMounts(
			"/fake-ephemeral-path",
			"/fake-persistent-dir",
			logger,
		)
	})

	Describe("MakeEphemeral", func() {
		It("returns ephemeral path", func() {
			path := guestBindMounts.MakeEphemeral()
			Expect(path).To(Equal("/fake-ephemeral-path"))
		})
	})

	Describe("MakePersistent", func() {
		It("returns persistent dir", func() {
			path := guestBindMounts.MakePersistent()
			Expect(path).To(Equal("/fake-persistent-dir"))
		})
	})

	Describe("MountPersistent", func() {
		It("returns persistent dir + disk id", func() {
			path := guestBindMounts.MountPersistent("fake-disk-id")
			Expect(path).To(Equal("/fake-persistent-dir/fake-disk-id"))
		})
	})
})
