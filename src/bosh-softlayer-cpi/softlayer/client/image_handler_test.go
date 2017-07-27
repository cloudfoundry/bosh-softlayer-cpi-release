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

var _ = Describe("ImageHandler", func() {
	var (
		server      *ghttp.Server
		vpsEndPoint string
		vps         *vpsVm.Client

		sess *session.Session
		cli  *boslc.ClientManager

		imageID int
	)
	BeforeEach(func() {
		// the fake server to setup VPS Server
		server = ghttp.NewServer()
		vpsEndPoint = server.URL()
		vps = vpsClient.New(httptransport.New(vpsEndPoint,
			"v2", []string{"http"}), strfmt.Default).VM

		sess = test_helpers.NewFakeSoftlayerSession(server)
		cli = boslc.NewSoftLayerClientManager(sess, vps)

		imageID = 1335057
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("GetImage", func() {
		Context("when ImageService getObject call successfully", func() {
			It("get image successfully", func() {
				image, succ, err := cli.GetImage(imageID, boslc.IMAGE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*image.Id).To(Equal(imageID))
			})
		})

		Context("when ImageService getObject call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetImage(imageID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

})
