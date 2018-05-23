package stemcell_test

import (
	"errors"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshfs "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	cpiLog "bosh-softlayer-cpi/logger"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	stemcellService "bosh-softlayer-cpi/softlayer/stemcell_service"
)

var _ = Describe("Stemcell Service", func() {
	var (
		err error

		stemcellID int
		imagePath  string
		datacenter string
		osCode     string

		cli        *fakeslclient.FakeClient
		stemcell   stemcellService.SoftlayerStemcellService
		uuidGen    *fakeuuid.FakeGenerator
		logger     cpiLog.Logger
		fs         boshsys.FileSystem
		cmdRunner  boshsys.CmdRunner
		compressor boshfs.Compressor
	)
	BeforeEach(func() {

		stemcellID = 22345678
		datacenter = "fake-datacenter"
		osCode = "fake-os-code"

		// create temp image tarball
		vhdFilePath := "/tmp/image.vhd"
		fs = boshsys.NewOsFileSystem(boshlog.New(boshlog.LevelDebug, log.New(os.Stdout, "", log.LstdFlags), log.New(os.Stderr, "", log.LstdFlags)))
		fs.WriteFileString(vhdFilePath, "")
		cmdRunner = boshsys.NewExecCmdRunner(boshlog.New(boshlog.LevelDebug, log.New(os.Stdout, "", log.LstdFlags), log.New(os.Stderr, "", log.LstdFlags)))
		defer func() {
			cmdRunner.RunCommand("rm", vhdFilePath)
		}()
		compressor = boshfs.NewTarballCompressor(cmdRunner, fs)
		imagePath, err = compressor.CompressSpecificFilesInDir("/tmp", []string{"image.vhd"})
		Expect(err).ToNot(HaveOccurred())

		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = cpiLog.NewLogger(boshlog.LevelDebug, "fake-thread-number-id")
		stemcell = stemcellService.NewSoftlayerStemcellService(cli, uuidGen, logger)

	})

	Describe("Call CreateFromTarball", func() {
		Context("when softlayerClient GetImage call successfully", func() {
			It("create stemcell successfully", func() {
				cli.CreateSwiftContainerReturns(
					nil,
				)
				cli.UploadSwiftLargeObjectReturns(
					nil,
				)
				cli.CreateImageFromExternalSourceReturns(
					stemcellID,
					nil,
				)
				cli.DeleteSwiftLargeObjectReturns(
					nil,
				)
				cli.DeleteSwiftContainerReturns(
					nil,
				)

				globalIdentifier, err := stemcell.CreateFromTarball(imagePath, datacenter, osCode)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CreateSwiftContainerCallCount()).To(Equal(1))
				Expect(cli.UploadSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.CreateImageFromExternalSourceCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftContainerCallCount()).To(Equal(1))
				Expect(globalIdentifier).To(Equal(stemcellID))

				cmdRunner.RunCommand("rm " + imagePath)
			})
		})

		Context("when softlayerClient CreateSwiftContainer call return error", func() {
			It("failed to create stemcell", func() {
				cli.CreateSwiftContainerReturns(
					errors.New("fake-client-error"),
				)

				_, err = stemcell.CreateFromTarball(imagePath, datacenter, osCode)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CreateSwiftContainerCallCount()).To(Equal(1))
				Expect(cli.UploadSwiftLargeObjectCallCount()).To(Equal(0))
				Expect(cli.CreateImageFromExternalSourceCallCount()).To(Equal(0))
				Expect(cli.DeleteSwiftLargeObjectCallCount()).To(Equal(0))
				Expect(cli.DeleteSwiftContainerCallCount()).To(Equal(1))
			})
		})

		Context("when softlayerClient UploadSwiftLargeObject call return error", func() {
			It("failed to create stemcell", func() {
				cli.CreateSwiftContainerReturns(
					nil,
				)
				cli.UploadSwiftLargeObjectReturns(
					errors.New("fake-client-error"),
				)
				cli.DeleteSwiftContainerReturns(
					nil,
				)

				_, err = stemcell.CreateFromTarball(imagePath, datacenter, osCode)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CreateSwiftContainerCallCount()).To(Equal(1))
				Expect(cli.UploadSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.CreateImageFromExternalSourceCallCount()).To(Equal(0))
				Expect(cli.DeleteSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftContainerCallCount()).To(Equal(1))
			})
		})

		Context("when softlayerClient CreateImageFromExternalSource call return error", func() {
			It("failed to create stemcell", func() {
				cli.CreateSwiftContainerReturns(
					nil,
				)
				cli.UploadSwiftLargeObjectReturns(
					nil,
				)
				cli.CreateImageFromExternalSourceReturns(
					0,
					errors.New("fake-client-error"),
				)
				cli.DeleteSwiftLargeObjectReturns(
					nil,
				)
				cli.DeleteSwiftContainerReturns(
					nil,
				)

				globalIdentifier, err := stemcell.CreateFromTarball(imagePath, datacenter, osCode)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CreateSwiftContainerCallCount()).To(Equal(1))
				Expect(cli.UploadSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.CreateImageFromExternalSourceCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftContainerCallCount()).To(Equal(1))
				Expect(globalIdentifier).NotTo(Equal(stemcellID))
			})
		})

		Context("when softlayerClient DeleteSwiftLargeObject call return error", func() {
			It("create stemcell successfully and only print error in defer statement", func() {
				cli.CreateSwiftContainerReturns(
					nil,
				)
				cli.UploadSwiftLargeObjectReturns(
					nil,
				)
				cli.CreateImageFromExternalSourceReturns(
					stemcellID,
					nil,
				)
				cli.DeleteSwiftLargeObjectReturns(
					errors.New("fake-client-error"),
				)
				cli.DeleteSwiftContainerReturns(
					nil,
				)

				globalIdentifier, err := stemcell.CreateFromTarball(imagePath, datacenter, osCode)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CreateSwiftContainerCallCount()).To(Equal(1))
				Expect(cli.UploadSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.CreateImageFromExternalSourceCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftContainerCallCount()).To(Equal(1))
				Expect(globalIdentifier).To(Equal(stemcellID))
			})
		})

		Context("when softlayerClient DeleteSwiftContainer call return error", func() {
			It("create stemcell successfully and only print error in defer statement", func() {
				cli.CreateSwiftContainerReturns(
					nil,
				)
				cli.UploadSwiftLargeObjectReturns(
					nil,
				)
				cli.CreateImageFromExternalSourceReturns(
					stemcellID,
					nil,
				)
				cli.DeleteSwiftLargeObjectReturns(
					nil,
				)
				cli.DeleteSwiftContainerReturns(
					errors.New("fake-client-error"),
				)

				globalIdentifier, err := stemcell.CreateFromTarball(imagePath, datacenter, osCode)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CreateSwiftContainerCallCount()).To(Equal(1))
				Expect(cli.UploadSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.CreateImageFromExternalSourceCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftLargeObjectCallCount()).To(Equal(1))
				Expect(cli.DeleteSwiftContainerCallCount()).To(Equal(1))
				Expect(globalIdentifier).To(Equal(stemcellID))
			})
		})
	})

	AfterEach(func() {
		cmdRunner = boshsys.NewExecCmdRunner(boshlog.New(boshlog.LevelDebug, log.New(os.Stdout, "", log.LstdFlags), log.New(os.Stderr, "", log.LstdFlags)))
		defer func() {
			cmdRunner.RunCommand("rm", imagePath)
		}()
	})
})
