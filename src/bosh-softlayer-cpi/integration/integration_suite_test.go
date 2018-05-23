package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/ncw/swift"
	"github.com/softlayer/softlayer-go/datatypes"

	cfg "bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/softlayer/client"
	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	"github.com/softlayer/softlayer-go/filter"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Required env vars
	Expect(username).ToNot(Equal(""), "SL_USERNAME must be set")
	Expect(apiKey).ToNot(Equal(""), "SL_API_KEY must be set")

	// Setup Register HTTP server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()

		switch r.Method {
		case "PUT":
			w.WriteHeader(http.StatusCreated)
		case "DELETE":
			w.WriteHeader(http.StatusOK)
		}
	})
	ts = httptest.NewServer(handler)
	url, err := url.Parse(ts.URL)
	Expect(err).To(BeNil())
	registerPort, err := strconv.Atoi(strings.Split(url.Host, ":")[1])
	Expect(err).To(BeNil())

	// Update cfgContent with Endpoint info
	cfgContent = fmt.Sprintf(
		`{
		  "cloud": {
		    "plugin": "softlayer",
		    "properties": {
		      "softlayer": {
			  "username": "%s",
			  "api_key": "%s",
			  "swift_username": "%s",
			  "swift_endpoint": "%s"
		      },
		      "registry": {
			"user": "registry",
			"password": "1330c82d-4bc4-4544-4a90-c2c78fa66431",
			"address": "127.0.0.1",
			"http": {
			  "port": %d,
			  "user": "registry",
			  "password": "1330c82d-4bc4-4544-4a90-c2c78fa66431"
			},
			"endpoint": "%s"
		      },
		      "agent": {
			"ntp": [],
			"blobstore": {
			  "provider": "dav",
			  "options": {
			    "endpoint": "http://127.0.0.1:25250",
			    "user": "agent",
			    "password": "agent"
			  }
			},
			"mbus": "nats://nats:nats@127.0.0.1:4222"
		      }
		    }
		  }
		}`, username, apiKey, swiftUsername, swiftEndpoint, registerPort, ts.URL)
	config, err = cfg.NewConfigFromString(cfgContent)
	Expect(err).To(BeNil())

	// Initialize session of softlayer client
	sess = client.NewSoftlayerClientSession(client.SoftlayerAPIEndpointPublicDefault, username, apiKey, true, timeout, retries, retryTimeout, clientLogger)

	// Setup vps client
	if config.Cloud.Properties.SoftLayer.EnableVps {
		vps = vpsClient.New(httptransport.New(fmt.Sprintf("%s:%d", config.Cloud.Properties.SoftLayer.VpsHost, config.Cloud.Properties.SoftLayer.VpsPort),
			"v2", []string{"https"}), strfmt.Default).VM
	}

	// Clean existing vms for integration test
	cleanVMs()

	// Create stemcell for integration test
	request := fmt.Sprintf(`{
			  "method": "create_stemcell",
			  "arguments": ["%s", {
			    "virtual-disk-image-id": %s,
			    "virtual-disk-image-uuid": "%s",
			    "datacenter-name": "%s",
			    "infrastructure": "softlayer"
			  }]
			}`, stemcellFile, stemcellId, stemcellUuid, datacenter)

	stemcell := assertSucceedsWithResult(request).(string)

	ips = make(chan string, len(ipAddrs))

	// Parse IP addresses to be used and put on a chan
	for _, addr := range ipAddrs {
		ips <- addr
	}

	return []byte(stemcell)
}, func(data []byte) {
	// Ensure stemcell was initialized
	existingStemcellId = string(data)
	Expect(existingStemcellId).ToNot(BeEmpty())
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	cleanVMs()
	ts.Close()
})

func cleanVMs() {
	// Initialize a compute API client ImageService
	//Swift Object Storage
	var swiftClient *swift.Connection
	if config.Cloud.Properties.SoftLayer.SwiftEndpoint != "" {
		swiftClient = client.NewSwiftClient(config.Cloud.Properties.SoftLayer.SwiftEndpoint, config.Cloud.Properties.SoftLayer.SwiftUsername, config.Cloud.Properties.SoftLayer.ApiKey, 120, 3)
	}

	softlayerClient := client.NewSoftLayerClientManager(sess, vps, swiftClient, multiLogger)
	accountService := softlayerClient.AccountService

	// Clean up any VMs left behind from failed tests. Instances with the 'blusbosh-slcpi-integration-test' prefix will be deleted.
	toDelete := make([]datatypes.Virtual_Guest, 0)
	GinkgoWriter.Write([]byte("Looking for VMs with 'bluebosh-slcpi-integration-test' prefix. Matches will be deleted\n"))
	// Clean up VMs with 'bluebosh-slcpi-integration-test' prefix hostname
	filter := filter.Build(
		filter.Path("virtualGuests.hostname").Eq("^= bluebosh-slcpi-integration-test"),
	)
	vgs, err := accountService.Filter(filter).Mask("id, hostname, fullyQualifiedDomainName").GetVirtualGuests()
	Expect(err).To(BeNil())
	for _, instance := range vgs {
		if strings.Contains(*instance.Hostname, "bluebosh-slcpi-integration-test") {
			toDelete = append(toDelete, instance)
		}
	}

	for _, vm := range toDelete {
		vmStatus, err := softlayerClient.VirtualGuestService.Id(int(*vm.Id)).GetStatus()
		Expect(err).To(BeNil())
		if *vmStatus.KeyName != "DISCONNECTED" {
			GinkgoWriter.Write([]byte(fmt.Sprintf("Deleting VM %s \n", *vm.FullyQualifiedDomainName)))
			//_, err := softlayerClient.VirtualGuestService.Id(int(*vm.Id)).DeleteObject()
			//Expect(err).ToNot(HaveOccurred())
		}
	}
}
