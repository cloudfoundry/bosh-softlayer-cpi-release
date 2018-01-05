package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/ncw/swift"

	"bosh-softlayer-cpi/action"
	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/api/dispatcher"
	"bosh-softlayer-cpi/api/transport"
	"bosh-softlayer-cpi/config"
	cpiLog "bosh-softlayer-cpi/logger"
	"bosh-softlayer-cpi/softlayer/client"
	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	"bosh-softlayer-cpi/softlayer/vps_service/client/vm"
)

const logTagMain = "main"

var (
	configPathOpt = flag.String("configFile", "", "Path to configuration file")
)

func main() {
	logger, fs, uuid, outLogger := basicDeps()
	cmdRunner := boshsys.NewExecCmdRunner(logger.GetBoshLogger())
	defer logger.HandlePanic("Main")

	flag.Parse()

	cfg, err := config.NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error(logTagMain, "Loading config %s", err.Error())
		os.Exit(1)
	}

	dispatch := buildDispatcher(cfg, logger, outLogger, uuid, cmdRunner)

	cli := transport.NewCLI(os.Stdin, os.Stdout, dispatch, logger)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error(logTagMain, "Serving once %s", err)
		os.Exit(1)
	}
}

func basicDeps() (cpiLog.Logger, boshsys.FileSystem, boshuuid.Generator, *log.Logger) {
	var logBuff bytes.Buffer
	multiWriter := io.MultiWriter(os.Stderr, bufio.NewWriter(&logBuff))
	nanos := fmt.Sprintf("%09d", time.Now().Nanosecond())

	clientLogger := log.New(multiWriter, nanos, log.LstdFlags) // For softlayer_client
	outLogger := log.New(multiWriter, "", log.LstdFlags)
	errLogger := log.New(os.Stderr, "", log.LstdFlags)

	cpiLogger := cpiLog.New(boshlog.LevelDebug, nanos, outLogger, errLogger)
	multiLogger := api.MultiLogger{Logger: cpiLogger, LogBuff: &logBuff}
	fs := boshsys.NewOsFileSystem(cpiLogger.GetBoshLogger())

	uuidGen := boshuuid.NewGenerator()

	return multiLogger, fs, uuidGen, clientLogger
}

func buildDispatcher(
	config config.Config,
	logger cpiLog.Logger,
	outLogger *log.Logger,
	uuidGen boshuuid.Generator,
	cmdRunner boshsys.CmdRunner,
) dispatcher.Dispatcher {
	var softlayerAPIEndpoint string
	if config.Cloud.Properties.SoftLayer.ApiEndpoint != "" {
		softlayerAPIEndpoint = config.Cloud.Properties.SoftLayer.ApiEndpoint
	} else {
		softlayerAPIEndpoint = client.SoftlayerAPIEndpointPublicDefault
	}

	softLayerClient := client.NewSoftlayerClientSession(softlayerAPIEndpoint, config.Cloud.Properties.SoftLayer.Username, config.Cloud.Properties.SoftLayer.ApiKey, true, 300, 3, 60, outLogger)

	var vps *vm.Client
	if config.Cloud.Properties.SoftLayer.EnableVps {
		vps = vpsClient.New(httptransport.New(fmt.Sprintf("%s:%d", config.Cloud.Properties.SoftLayer.VpsHost, config.Cloud.Properties.SoftLayer.VpsPort),
			"v2", []string{"https"}), strfmt.Default).VM
	}

	//Swift Object Storage
	var swiftClient *swift.Connection
	if config.Cloud.Properties.SoftLayer.SwiftEndpoint != "" {
		swiftClient = client.NewSwiftClient(config.Cloud.Properties.SoftLayer.SwiftEndpoint, config.Cloud.Properties.SoftLayer.SwiftUsername, config.Cloud.Properties.SoftLayer.ApiKey, 120, 3)
	}

	repClientFactory := client.NewClientFactory(client.NewSoftLayerClientManager(softLayerClient, vps, swiftClient, logger))
	cli := repClientFactory.CreateClient()

	actionFactory := action.NewConcreteFactory(
		cli,
		uuidGen,
		config,
		logger,
	)

	caller := dispatcher.NewJSONCaller()

	return dispatcher.NewJSON(actionFactory, caller, logger)
}
