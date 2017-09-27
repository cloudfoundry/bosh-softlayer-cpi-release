package snapshot_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/snapshot_service"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("Snapshot Service", func() {
	var (
		cli             *fakeslclient.FakeClient
		uuidGen         *fakeuuid.FakeGenerator
		logger          boshlog.Logger
		snapshotService SoftlayerSnapshotService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		snapshotService = NewSoftlayerSnapshotService(cli, logger)
	})

	Describe("Call Create", func() {
		var (
			diskID int
		)

		BeforeEach(func() {
			diskID = 12345678
		})

		It("Create successfully", func() {
			cli.CreateSnapshotReturns(
				datatypes.Network_Storage{
					Id: sl.Int(2345),
				},
				nil,
			)

			_, err := snapshotService.Create(diskID, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.CreateSnapshotCallCount()).To(Equal(1))
		})

		It("Create successfully", func() {
			cli.CreateSnapshotReturns(
				datatypes.Network_Storage{
					Id: sl.Int(2345),
				},
				nil,
			)

			_, err := snapshotService.Create(diskID, "note-string")
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.CreateSnapshotCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient CreateSnapshot call returns an error", func() {
			cli.CreateSnapshotReturns(
				datatypes.Network_Storage{},
				errors.New("fake-client-error"),
			)

			_, err := snapshotService.Create(diskID, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.CreateSnapshotCallCount()).To(Equal(1))
		})
	})
})
