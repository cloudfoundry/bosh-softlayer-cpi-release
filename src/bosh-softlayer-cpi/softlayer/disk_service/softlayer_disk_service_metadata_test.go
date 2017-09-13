package disk_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "bosh-softlayer-cpi/softlayer/disk_service"
	"fmt"
)

var _ = Describe("Virtual Guest Service", func() {
	var (
		cli         *fakeslclient.FakeClient
		logger      boshlog.Logger
		diskService SoftlayerDiskService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		diskService = NewSoftlayerDiskService(cli, logger)
	})

	Describe("Call SetMetadata", func() {
		var (
			diskID   int
			metaData Metadata
		)

		BeforeEach(func() {
			diskID = 12345678
			metaData = Metadata{
				"director":       "bats-director",
				"deployment":     "automation-cf",
				"instance_id":    "1f981b7e-1663-4aeb-8b7c-5e7d8783a4e3",
				"job":            "consul",
				"instance_index": "0",
				"instance_name":  "consul/1f981b7e-1663-4aeb-8b7c-5e7d8783a4e3",
				"attached_at":    "2017-09-12T14:43:41Z",
			}
		})

		It("Set notes successfully", func() {
			cli.SetNotesReturns(
				true,
				nil,
			)

			err := diskService.SetMetadata(diskID, metaData)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.SetNotesCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient SetNotes call returns an error", func() {
			cli.SetNotesReturns(
				false,
				errors.New("fake-client-error"),
			)

			err := diskService.SetMetadata(diskID, metaData)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.SetNotesCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient SetNotes call returns an error", func() {
			cli.SetNotesReturns(
				false,
				nil,
			)

			err := diskService.SetMetadata(diskID, metaData)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Disk '%d' not found", diskID)))
			Expect(cli.SetNotesCallCount()).To(Equal(1))
		})
	})
})
