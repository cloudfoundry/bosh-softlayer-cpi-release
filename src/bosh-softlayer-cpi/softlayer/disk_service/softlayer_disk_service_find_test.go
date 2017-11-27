package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	cpiLog "bosh-softlayer-cpi/logger"
	boslc "bosh-softlayer-cpi/softlayer/client/fakes"
	diskService "bosh-softlayer-cpi/softlayer/disk_service"
)

var _ = Describe("Disk Service Find", func() {
	var (
		err error

		diskID int
		cli    *boslc.FakeClient
		disk   diskService.SoftlayerDiskService
		logger cpiLog.Logger
	)
	BeforeEach(func() {
		diskID = 12345678
		cli = &boslc.FakeClient{}
		logger = cpiLog.NewLogger(boshlog.LevelDebug, "")
		disk = diskService.NewSoftlayerDiskService(cli, logger)
	})

	Describe("Call Find", func() {
		Context("when softlayerClient GetBlockVolumeDetails call successfully", func() {
			It("find delete successfully", func() {
				cli.GetBlockVolumeDetailsReturns(
					&datatypes.Network_Storage{
						Id: sl.Int(12345678),
					},
					true,
					nil,
				)

				networkStorage, err := disk.Find(diskID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetBlockVolumeDetailsCallCount()).To(Equal(1))
				Expect(*networkStorage.Id).To(Equal(diskID))
			})
		})

		Context("return error when softlayerClient GetBlockVolumeDetails call return error", func() {
			It("failed to find volume", func() {
				cli.GetBlockVolumeDetailsReturns(
					&datatypes.Network_Storage{},
					false,
					errors.New("fake-client-error"),
				)

				_, err = disk.Find(diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.GetBlockVolumeDetailsCallCount()).To(Equal(1))
			})
		})

		Context("return error when softlayerClient GetBlockVolumeDetails call find nothing", func() {
			It("failed to find volume", func() {
				cli.GetBlockVolumeDetailsReturns(
					&datatypes.Network_Storage{},
					false,
					nil,
				)

				_, err = disk.Find(diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(cli.GetBlockVolumeDetailsCallCount()).To(Equal(1))
			})
		})
	})
})
