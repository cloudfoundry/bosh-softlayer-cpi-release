package test_helpers

import (
	. "github.com/onsi/gomega"

	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	"bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func SetupVPS(handler http.HandlerFunc) *vm.Client {
	// the fake server to setup VPS Server
	ts := httptest.NewServer(handler)
	url, err := url.Parse(ts.URL)
	Expect(err).ToNot(HaveOccurred())
	vps := vpsClient.New(httptransport.New(url.Host,
		"v2", []string{"http"}), strfmt.Default).VM

	return vps
}
