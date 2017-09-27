package snapshot_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/snapshot_service"
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

	Describe("Call Delete", func() {
		var (
			diskID int
		)

		BeforeEach(func() {
			diskID = 12345678
		})

		It("Clean up successfully", func() {
			cli.DeleteSnapshotReturns(
				nil,
			)

			err := snapshotService.Delete(diskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.DeleteSnapshotCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient delete snapshot call returns an error", func() {
			cli.DeleteSnapshotReturns(
				errors.New("fake-client-error"),
			)

			err := snapshotService.Delete(diskID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.DeleteSnapshotCallCount()).To(Equal(1))
		})
	})
})
