package main

import (
	"flag"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-softlayer-cpi/action"
	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/api/dispatcher"
	"bosh-softlayer-cpi/api/transport"
	"bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/softlayer/client"
	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	"bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bufio"
	"bytes"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"io"
)

const logTagMain = "main"

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
	cpiVersion    = flag.Bool("version", false, "The version of CPI release")
)

func main() {
	logger, fs, uuid, writer := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	if *cpiVersion {
		os.Exit(0)
	}

	cfg, err := config.NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error(logTagMain, "Loading config %s", err.Error())
		os.Exit(1)
	}

	dispatch := buildDispatcher(cfg, logger, writer, uuid)

	cli := transport.NewCLI(os.Stdin, os.Stdout, dispatch, logger)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error(logTagMain, "Serving once %s", err)
		os.Exit(1)
	}
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem, boshuuid.Generator, io.Writer) {
	var logBuff bytes.Buffer
	multiWriter := io.MultiWriter(os.Stderr, bufio.NewWriter(&logBuff))
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, multiWriter, os.Stderr)
	multiLogger := api.MultiLogger{Logger: logger, LogBuff: &logBuff}
	fs := boshsys.NewOsFileSystem(multiLogger)

	uuidGen := boshuuid.NewGenerator()

	return multiLogger, fs, uuidGen, multiWriter
}

func buildDispatcher(
	config config.Config,
	logger boshlog.Logger,
	writer io.Writer,
	uuidGen boshuuid.Generator,
) dispatcher.Dispatcher {
	var softlayerAPIEndpoint string
	if config.Cloud.Properties.SoftLayer.ApiEndpoint != "" {
		softlayerAPIEndpoint = config.Cloud.Properties.SoftLayer.ApiEndpoint
	} else {
		softlayerAPIEndpoint = client.SoftlayerAPIEndpointPublicDefault
	}

	softLayerClient := client.NewSoftlayerClientSession(softlayerAPIEndpoint, config.Cloud.Properties.SoftLayer.Username, config.Cloud.Properties.SoftLayer.ApiKey, true, 300, writer)

	var vps *vm.Client
	if config.Cloud.Properties.SoftLayer.EnableVps {
		if config.Cloud.Properties.SoftLayer.VpsUseSsl {
			vps = vpsClient.New(httptransport.New(fmt.Sprintf("%s:%d", config.Cloud.Properties.SoftLayer.VpsHost, config.Cloud.Properties.SoftLayer.VpsPort),
				"v2", []string{"https"}), strfmt.Default).VM
		} else {

			vps = vpsClient.New(httptransport.New(fmt.Sprintf("%s:%d", config.Cloud.Properties.SoftLayer.VpsHost, config.Cloud.Properties.SoftLayer.VpsPort),
				"v2", []string{"http"}), strfmt.Default).VM
		}

	}

	repClientFactory := client.NewClientFactory(client.NewSoftLayerClientManager(softLayerClient, vps))
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
