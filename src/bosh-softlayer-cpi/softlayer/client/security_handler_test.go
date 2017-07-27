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
)

var _ = Describe("SecurityHandler", func() {
	var (
		server      *ghttp.Server
		vpsEndPoint string
		vps         *vpsVm.Client

		sess *session.Session
		cli  *boslc.ClientManager

		label       string
		key         string
		fingerPrint string
		sshKeyId    int
	)
	BeforeEach(func() {
		// the fake server to setup VPS Server
		server = ghttp.NewServer()
		vpsEndPoint = server.URL()
		vps = vpsClient.New(httptransport.New(vpsEndPoint,
			"v2", []string{"http"}), strfmt.Default).VM

		sess = test_helpers.NewFakeSoftlayerSession(server)
		cli = boslc.NewSoftLayerClientManager(sess, vps)

		label = "fake-label"
		key = "fake-key"
		fingerPrint = "fakte-fingerPrint"
		sshKeyId = 12345678
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("CreateSshKey", func() {
		Context("when ImageService createObject call successfully", func() {
			It("create ssh key successfully", func() {
				_, err := cli.CreateSshKey(&label, &key, &fingerPrint)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteSshKey", func() {
		Context("when ImageService deleteObject call successfully", func() {
			It("delete ssh key successfully", func() {
				succeed, err := cli.DeleteSshKey(sshKeyId)
				Expect(err).NotTo(HaveOccurred())
				Expect(succeed).To(Equal(true))
			})
		})
	})

})
