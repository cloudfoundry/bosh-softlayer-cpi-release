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

var _ = Describe("SecurityHandler", func() {
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

		label       string
		key         string
		fingerPrint string
		sshKeyId    int

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

		label = "fake-label"
		key = "fake-key"
		fingerPrint = "fake-fingerPrint"
		sshKeyId = 12345678
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("CreateSshKey", func() {
		Context("when SoftLayerSecuritySshKey createObject call successfully", func() {
			It("Create ssh key successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Security_Ssh_Key_createObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateSshKey(&label, &key, &fingerPrint)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Create ssh key successfully when vgs has sshkeys", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Security_Ssh_Key_createObject_PublicException.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Account_getSshKeys.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateSshKey(&label, &key, &fingerPrint)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when SoftLayerSecuritySshKey createObject call return an error", func() {
			It("Return error when ssh key successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Security_Ssh_Key_createObject_PublicException.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Account_getSshKeys_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateSshKey(&label, &key, &fingerPrint)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})
		})
	})

	Describe("DeleteSshKey", func() {
		Context("when SoftLayerSecuritySshKey deleteObject call successfully", func() {
			It("Delete ssh key successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Security_Ssh_Key_deleteObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succeed, err := cli.DeleteSshKey(sshKeyId)
				Expect(err).NotTo(HaveOccurred())
				Expect(succeed).To(Equal(true))
			})
		})

		Context("when SoftLayerSecuritySshKey deleteObject call return an error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Security_Ssh_Key_deleteObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succeed, err := cli.DeleteSshKey(sshKeyId)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succeed).To(Equal(false))
			})
		})
	})
})
