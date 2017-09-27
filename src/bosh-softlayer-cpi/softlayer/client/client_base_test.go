package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boslc "bosh-softlayer-cpi/softlayer/client"
	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	"bosh-softlayer-cpi/test_helpers"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("ClientBaseTest", func() {
	It("Create client", func() {
		server := ghttp.NewServer()
		vpsEndPoint := server.URL()
		vps := vpsClient.New(httptransport.New(vpsEndPoint,
			"v2", []string{"http"}), strfmt.Default).VM

		transportHandler := &test_helpers.FakeTransportHandler{
			FakeServer:           server,
			SoftlayerAPIEndpoint: server.URL(),
			MaxRetries:           3,
		}
		sess := test_helpers.NewFakeSoftlayerSession(transportHandler)
		cli := boslc.NewSoftLayerClientManager(sess, vps)

		cliFactory := boslc.NewClientFactory(cli)
		returnedCli := cliFactory.CreateClient()

		Expect(returnedCli).To(Equal(cli))
		defer func() {
			// Inspect no panic occur
			r := recover()
			Expect(r).To(BeNil())
		}()
	})
})
