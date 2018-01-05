package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"strconv"
	"time"

	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/ncw/swift"
	"github.com/onsi/gomega/ghttp"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("ClientBaseTest", func() {
	It("Create client", func() {
		server := ghttp.NewServer()
		transportHandler := &test_helpers.FakeTransportHandler{
			FakeServer:           server,
			SoftlayerAPIEndpoint: server.URL(),
			MaxRetries:           3,
		}

		vps := &vpsVm.Client{}
		swiftClient := &swift.Connection{}

		var errOutLog bytes.Buffer
		nanos := time.Now().Nanosecond()
		logger := logger.NewLogger(boshlogger.LevelDebug, strconv.Itoa(nanos))
		multiLogger := api.MultiLogger{Logger: logger, LogBuff: &errOutLog}
		sess := test_helpers.NewFakeSoftlayerSession(transportHandler)
		cli := slClient.NewSoftLayerClientManager(sess, vps, swiftClient, multiLogger)

		cliFactory := slClient.NewClientFactory(cli)
		returnedCli := cliFactory.CreateClient()

		Expect(returnedCli).To(Equal(cli))
		defer func() {
			// Inspect no panic occur
			r := recover()
			Expect(r).To(BeNil())
		}()
	})
})
