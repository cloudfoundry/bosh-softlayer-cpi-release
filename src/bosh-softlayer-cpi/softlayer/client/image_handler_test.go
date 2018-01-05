package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"net/http"
	"strconv"
	"time"

	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/ncw/swift"
	"github.com/onsi/gomega/ghttp"
	"github.com/softlayer/softlayer-go/session"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("ImageHandler", func() {
	var (
		err error

		errOutLog   bytes.Buffer
		logger      cpiLog.Logger
		multiLogger api.MultiLogger

		server      *ghttp.Server
		vps         *vpsVm.Client
		swiftClient *swift.Connection

		transportHandler *test_helpers.FakeTransportHandler
		sess             *session.Session
		cli              *slClient.ClientManager

		imageID   int
		respParas []map[string]interface{}
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
		transportHandler = &test_helpers.FakeTransportHandler{
			FakeServer:           server,
			SoftlayerAPIEndpoint: server.URL(),
			MaxRetries:           3,
		}

		vps = &vpsVm.Client{}
		swiftClient = &swift.Connection{}

		nanos := time.Now().Nanosecond()
		logger = cpiLog.NewLogger(boshlogger.LevelDebug, strconv.Itoa(nanos))
		multiLogger = api.MultiLogger{Logger: logger, LogBuff: &errOutLog}
		sess = test_helpers.NewFakeSoftlayerSession(transportHandler)
		cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

		imageID = 1335057
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("GetImage", func() {
		Context("when ImageService getObject call successfully", func() {
			It("get image successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				image, succ, err := cli.GetImage(imageID, slClient.IMAGE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*image.Id).To(Equal(imageID))
			})

			It("get image successfully when pass empty mask", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				image, succ, err := cli.GetImage(imageID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*image.Id).To(Equal(imageID))
			})
		})

		Context("when ImageService getObject call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, succ, err := cli.GetImage(imageID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})

			It("return an empty image when ImageService getObject call return empty object", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, succ, err := cli.GetImage(imageID, slClient.IMAGE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(false))
			})
		})
	})

})
