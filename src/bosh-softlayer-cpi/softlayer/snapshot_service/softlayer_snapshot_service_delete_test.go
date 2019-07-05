package snapshot_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cpiLog "bosh-softlayer-cpi/logger"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	. "bosh-softlayer-cpi/softlayer/snapshot_service"
	boshlog "github.com/bluebosh/bosh-utils/logger"
)

var _ = Describe("Snapshot Service", func() {
	var (
		cli             *fakeslclient.FakeClient
		logger          cpiLog.Logger
		snapshotService SoftlayerSnapshotService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		logger = cpiLog.NewLogger(boshlog.LevelDebug, "")
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
