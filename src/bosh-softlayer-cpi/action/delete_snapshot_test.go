package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	snapshotfakes "bosh-softlayer-cpi/softlayer/snapshot_service/fakes"
)

var _ = Describe("DeleteSnapshot", func() {
	var (
		err error

		snapshotCID SnapshotCID

		snapshotService *snapshotfakes.FakeService

		deleteSnapshot DeleteSnapshot
	)

	BeforeEach(func() {
		snapshotCID = SnapshotCID(22345678)
		snapshotService = &snapshotfakes.FakeService{}
		deleteSnapshot = NewDeleteSnapshot(snapshotService)
	})

	Describe("Run", func() {
		It("deletes the snapshot", func() {
			_, err = deleteSnapshot.Run(snapshotCID)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapshotService.DeleteCallCount()).To(Equal(1))
		})

		It("returns an error if snapshotService delete call returns an error", func() {
			snapshotService.DeleteReturns(errors.New("fake-snapshot-service-error"))
			_, err = deleteSnapshot.Run(snapshotCID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-snapshot-service-error"))
			Expect(snapshotService.DeleteCallCount()).To(Equal(1))
		})
	})
})
