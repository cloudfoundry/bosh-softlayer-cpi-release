package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/ncw/swift"
	"github.com/softlayer/softlayer-go/session"

	"bosh-softlayer-cpi/action"
	"bosh-softlayer-cpi/api"
	disp "bosh-softlayer-cpi/api/dispatcher"
	"bosh-softlayer-cpi/api/transport"
	cfg "bosh-softlayer-cpi/config"
	cpiLog "bosh-softlayer-cpi/logger"
	"bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bufio"
	"io"
	"log"
)

var (
	// A stemcell that will be created/loaded in integration_suite_test.go
	existingStemcellId string
	in, out, logBuffer bytes.Buffer
	username           = envRequired("SL_USERNAME")
	apiKey             = envRequired("SL_API_KEY")

	// Configurable defaults
	stemcellId   = envOrDefault("STEMCELL_ID", "1633205") // light-bosh-stemcell-3363.24.3-softlayer-xen-ubuntu-trusty-go_agent
	stemcellFile = envOrDefault("STEMCELL_FILE", "")
	stemcellUuid = envOrDefault("STEMCELL_UUID", "ea065435-f7ec-4f1c-8f3f-2987086b1427")
	datacenter   = envOrDefault("DATACENTER", "lon02")
	ipAddrs      = strings.Split(envOrDefault("PRIVATE_IP", "192.168.100.102,192.168.100.103,192.168.100.104"), ",")

	swiftUsername      = envOrDefault("SWIFT_USERNAME", "")
	swiftEndpoint      = envOrDefault("SWIFT_ENDPOINT", "")

	ts           *httptest.Server
	config       cfg.Config
	boshResponse disp.Response
	vps          *vpsVm.Client

	// Channel that will be used to retrieve IPs to use
	ips chan string

	trace        = false
	timeout      = 120
	retries      = 2
	retryTimeout = 60

	cfgContent = ""

	// Stuff of softlayer client
	multiWriter  = io.MultiWriter(os.Stderr, bufio.NewWriter(&logBuffer))
	clientLogger = log.New(multiWriter, "Integration", log.LstdFlags)
	outLogger    = log.New(multiWriter, "", log.LstdFlags)
	errLogger    = log.New(os.Stderr, "", log.LstdFlags)

	cpiLogger       = cpiLog.New(boshlog.LevelDebug, "Integration", outLogger, errLogger)
	multiLogger     = api.MultiLogger{Logger: cpiLogger, LogBuff: &logBuffer}
	uuidGen         = uuid.NewGenerator()
	sess            *session.Session
	softlayerClient = &client.ClientManager{}
)

func execCPI(request string) (disp.Response, error) {
	var err error

	//Swift Object Storage
	var swiftClient *swift.Connection

	if config.Cloud.Properties.SoftLayer.SwiftEndpoint != "" {
		swiftClient = client.NewSwiftClient(config.Cloud.Properties.SoftLayer.SwiftEndpoint, config.Cloud.Properties.SoftLayer.SwiftUsername, config.Cloud.Properties.SoftLayer.ApiKey, 120, 3)
	}

	softlayerClient = client.NewSoftLayerClientManager(sess, vps, swiftClient, multiLogger)
	actionFactory := action.NewConcreteFactory(
		softlayerClient,
		uuidGen,
		config,
		multiLogger,
	)

	caller := disp.NewJSONCaller()
	dispatcher := disp.NewJSON(actionFactory, caller, multiLogger)

	in.WriteString(request)
	cli := transport.NewCLI(&in, &out, dispatcher, multiLogger)

	var response []byte

	if err = cli.ServeOnce(); err != nil {
		return boshResponse, err
	}

	if response, err = ioutil.ReadAll(&out); err != nil {
		return boshResponse, err
	}

	if err = json.Unmarshal(response, &boshResponse); err != nil {
		return boshResponse, err
	}
	return boshResponse, nil
}

func envRequired(key string) (val string) {
	if val = os.Getenv(key); val == "" {
		panic(fmt.Sprintf("Could not find required environment variable '%s'", key))
	}
	return
}

func envOrDefault(key, defaultVal string) (val string) {
	if val = os.Getenv(key); val == "" {
		val = defaultVal
	}
	return
}
