package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boslc "bosh-softlayer-cpi/softlayer/client"
	"bosh-softlayer-cpi/test_helpers"
	"github.com/go-openapi/strfmt"
	"github.com/onsi/gomega/ghttp"

	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/softlayer/softlayer-go/session"
	"net/http"
)

var _ = Describe("SecurityHandler", func() {
	var (
		err error

		server      *ghttp.Server
		vpsEndPoint string
		vps         *vpsVm.Client

		transportHandler *test_helpers.FakeTransportHandler
		sess             *session.Session
		cli              *boslc.ClientManager

		label       string
		key         string
		fingerPrint string
		sshKeyId    int

		respParas []map[string]interface{}
	)
	BeforeEach(func() {
		// the fake server to setup VPS Server
		server = ghttp.NewServer()
		vpsEndPoint = server.URL()
		vps = vpsClient.New(httptransport.New(vpsEndPoint,
			"v2", []string{"http"}), strfmt.Default).VM

		transportHandler = &test_helpers.FakeTransportHandler{
			FakeServer:           server,
			SoftlayerAPIEndpoint: server.URL(),
			MaxRetries:           3,
		}
		sess = test_helpers.NewFakeSoftlayerSession(transportHandler)
		cli = boslc.NewSoftLayerClientManager(sess, vps)

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
