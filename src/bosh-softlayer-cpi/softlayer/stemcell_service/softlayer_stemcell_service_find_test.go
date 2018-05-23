package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	cpiLog "bosh-softlayer-cpi/logger"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	stemcellService "bosh-softlayer-cpi/softlayer/stemcell_service"
)

var _ = Describe("Stemcell Service", func() {
	var (
		err error

		stemcellID int
		cli        *fakeslclient.FakeClient
		stemcell   stemcellService.SoftlayerStemcellService
		uuidGen    *fakeuuid.FakeGenerator
		logger     cpiLog.Logger
	)
	BeforeEach(func() {
		stemcellID = 22345678
		cli = &fakeslclient.FakeClient{}
		logger = cpiLog.NewLogger(boshlog.LevelDebug, "")
		uuidGen = &fakeuuid.FakeGenerator{}
		stemcell = stemcellService.NewSoftlayerStemcellService(cli, uuidGen, logger)
	})

	Describe("Call Find", func() {
		Context("when softlayerClient GetImage call successfully", func() {
			It("find delete successfully", func() {
				cli.GetImageReturns(
					&datatypes.Virtual_Guest_Block_Device_Template_Group{
						GlobalIdentifier: sl.String("07beadaa-1e11-476e-a188-3f7795feb9fb"),
					},
					true,
					nil,
				)

				globalIdentifier, err := stemcell.Find(stemcellID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetImageCallCount()).To(Equal(1))
				Expect(globalIdentifier).To(Equal("07beadaa-1e11-476e-a188-3f7795feb9fb"))

			})
		})

		Context("return error when softlayerClient GetImage call return error", func() {
			It("failed to find volume", func() {
				cli.GetImageReturns(
					&datatypes.Virtual_Guest_Block_Device_Template_Group{},
					false,
					errors.New("fake-client-error"),
				)

				_, err = stemcell.Find(stemcellID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.GetImageCallCount()).To(BeNumerically(">=", 1))
			})
		})

		Context("return error when softlayerClient GetImage call find nothing", func() {
			It("failed to find volume", func() {
				cli.GetImageReturns(
					&datatypes.Virtual_Guest_Block_Device_Template_Group{},
					false,
					nil,
				)

				_, err = stemcell.Find(stemcellID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(cli.GetImageCallCount()).To(Equal(1))
			})
		})
	})

})
