package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	diskfakes "bosh-softlayer-cpi/softlayer/disk_service/fakes"
	snapshotfakes "bosh-softlayer-cpi/softlayer/snapshot_service/fakes"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("SnapshotDisk", func() {
	var (
		err        error
		metadata   SnapshotMetadata
		snapshotID string

		diskService     *diskfakes.FakeService
		snapshotService *snapshotfakes.FakeService

		snapshotDisk SnapshotDisk

		snapshotCID SnapshotCID
		diskCID     DiskCID
	)

	BeforeEach(func() {
		diskService = &diskfakes.FakeService{}
		snapshotService = &snapshotfakes.FakeService{}
		snapshotDisk = NewSnapshotDisk(snapshotService, diskService)

		snapshotCID = SnapshotCID(1234567)
		diskCID = DiskCID(2345678)
	})

	Describe("Run", func() {
		BeforeEach(func() {
			diskService.FindReturns(
				&datatypes.Network_Storage{
					Id: sl.Int(2345678),
				},
				nil,
			)
			snapshotService.CreateReturns(1234567, nil)
			metadata = SnapshotMetadata{Deployment: "fake-deployment", Job: "fake-job", Index: "fake-index"}
		})

		Context("creates a snaphot", func() {
			It("with the proper note", func() {
				snapshotID, err = snapshotDisk.Run(diskCID, metadata)
				Expect(err).NotTo(HaveOccurred())
				Expect(diskService.FindCallCount()).To(Equal(1))
				Expect(snapshotService.CreateCallCount()).To(Equal(1))
				Expect(snapshotID).To(Equal("1234567"))
			})

		})

		It("returns an error if diskService find call returns an error", func() {
			diskService.FindReturns(nil, errors.New("fake-disk-service-error"))

			_, err = snapshotDisk.Run(diskCID, metadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-disk-service-error"))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(snapshotService.CreateCallCount()).To(Equal(0))
		})

		It("returns an error if snapshotService create call returns an error", func() {
			snapshotService.CreateReturns(0, errors.New("fake-snapshot-service-error"))

			_, err = snapshotDisk.Run(diskCID, metadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-snapshot-service-error"))
			Expect(diskService.FindCallCount()).To(Equal(1))
			Expect(snapshotService.CreateCallCount()).To(Equal(1))
		})
	})
})
