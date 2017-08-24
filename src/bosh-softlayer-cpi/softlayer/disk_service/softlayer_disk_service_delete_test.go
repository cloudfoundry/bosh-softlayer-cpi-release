package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	diskService "bosh-softlayer-cpi/softlayer/disk_service"
	"errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var _ = Describe("Disk Service Delete", func() {
	var (
		err error

		diskID int
		cli    *fakeslclient.FakeClient
		disk   diskService.SoftlayerDiskService
		logger boshlog.Logger
	)
	BeforeEach(func() {
		diskID = 2048
		cli = &fakeslclient.FakeClient{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		disk = diskService.NewSoftlayerDiskService(cli, logger)
	})

	Describe("Call Delete", func() {
		Context("when softlayerClient CancelBlockVolume call successfully", func() {
			It("delete volume successfully", func() {
				cli.CancelBlockVolumeReturns(
					true,
					nil,
				)

				err = disk.Delete(diskID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CancelBlockVolumeCallCount()).To(Equal(1))
			})
		})

		Context("return error when softlayerClient CancelBlockVolume call return error", func() {
			It("failed to delete volume", func() {
				cli.CancelBlockVolumeReturns(
					false,
					errors.New("fake-client-error"),
				)

				err = disk.Delete(diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CancelBlockVolumeCallCount()).To(Equal(1))
			})
		})

		Context("return error when softlayerClient CancelBlockVolume call return billing item of volume is found", func() {
			It("failed to delete volume", func() {
				cli.CancelBlockVolumeReturns(
					false,
					errors.New("No billing item is found to cancel"),
				)

				err = disk.Delete(diskID)
				Expect(err).ToNot(HaveOccurred())
				Expect(cli.CancelBlockVolumeCallCount()).To(Equal(1))
			})
		})
	})
})
