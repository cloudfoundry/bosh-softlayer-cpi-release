package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	fakedisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk/fakes"
)

var _ = Describe("CreateDisk", func() {
	var (
		diskCreator *fakedisk.FakeCreator
		action      CreateDisk
	)

	BeforeEach(func() {
		diskCreator = &fakedisk.FakeCreator{}
		action = NewCreateDisk(diskCreator)
	})

	Describe("Run", func() {
		It("returns id for created disk for specific size", func() {
			diskCreator.CreateDisk = fakedisk.NewFakeDisk(1234)

			id, err := action.Run(20, VMCID(1234))
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(DiskCID(1234)))

			Expect(diskCreator.CreateSize).To(Equal(20))
		})

		It("returns error if creating disk fails", func() {
			diskCreator.CreateErr = errors.New("fake-create-err")

			id, err := action.Run(20, VMCID(1234))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-err"))
			Expect(id).To(Equal(DiskCID(0)))
		})
	})
})
