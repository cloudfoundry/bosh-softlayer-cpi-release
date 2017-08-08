package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	diskService "bosh-softlayer-cpi/softlayer/disk_service"
	"errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("Disk Service Create", func() {
	var (
		err error

		size     int
		iops     int
		location string

		cli    *fakeslclient.FakeClient
		disk   diskService.SoftlayerDiskService
		logger boshlog.Logger
	)
	BeforeEach(func() {
		size = 2048
		iops = 5000
		location = "lon02"

		cli = &fakeslclient.FakeClient{}
		logger = boshlog.NewLogger(boshlog.LevelDebug)
		disk = diskService.NewSoftlayerDiskService(cli, logger)
	})

	Describe("Call Create", func() {
		Context("when softlayerClient CreateVolume call successfully", func() {
			It("create volume successfully", func() {
				cli.CreateVolumeReturns(
					&datatypes.Network_Storage{
						Id: sl.Int(12345678),
					},
					nil,
				)

				_, err = disk.Create(size, iops, location)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CreateVolumeCallCount()).To(Equal(1))
			})
		})

		Context("return error when softlayerClient CreateVolume call return error", func() {
			It("failed to create volume", func() {
				cli.CreateVolumeReturns(
					&datatypes.Network_Storage{},
					errors.New("fake-client-error"),
				)

				_, err = disk.Create(size, iops, location)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CreateVolumeCallCount()).To(Equal(1))
			})
		})
	})
})
