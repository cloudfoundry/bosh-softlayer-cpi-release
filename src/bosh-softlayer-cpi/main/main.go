package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
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
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const mainLogTag = "main"

var (
	configFileOpt = flag.String("configFile", "", "Path to configuration file")
	input         io.Reader
	output        io.Writer
)

func main() {
	logger, fs, cmdRunner, uuid := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	cfg, err := config.NewConfigFromPath(*configFileOpt, fs)
	if err != nil {
		logger.Error(mainLogTag, "Loading config - %s", err.Error())
		os.Exit(1)
	}

	dispatcher := buildDispatcher(cfg, logger, cmdRunner, uuid)

	cli := transport.NewCLI(os.Stdin, os.Stdout, dispatcher, logger)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error(mainLogTag, "Serving once %s", err)
		os.Exit(1)
	}
}

func basicDeps() (api.MultiLogger, boshsys.FileSystem, boshsys.CmdRunner, boshuuid.Generator) {
	var logBuff bytes.Buffer
	multiWriter := io.MultiWriter(os.Stderr, bufio.NewWriter(&logBuff))
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, multiWriter, os.Stderr)
	multiLogger := api.MultiLogger{Logger: logger, LogBuff: &logBuff}
	fs := boshsys.NewOsFileSystem(multiLogger)

	cmdRunner := boshsys.NewExecCmdRunner(multiLogger)

	uuidGen := boshuuid.NewGenerator()

	return multiLogger, fs, cmdRunner, uuidGen
}

func buildDispatcher(
	cfg config.Config,
	logger api.MultiLogger,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
) dispatcher.Dispatcher {
	var softlayerAPIEndpoint string
	if cfg.Cloud.Properties.SoftLayer.ApiEndpoint != "" {
		softlayerAPIEndpoint = cfg.Cloud.Properties.SoftLayer.ApiEndpoint
	} else {
		softlayerAPIEndpoint = client.SoftlayerAPIEndpointPublicDefault
	}

	softLayerClient := client.NewSoftlayerClientSession(softlayerAPIEndpoint, cfg.Cloud.Properties.SoftLayer.Username, cfg.Cloud.Properties.SoftLayer.ApiKey, true, 300, logger)

	var vps *vm.Client
	if cfg.Cloud.Properties.SoftLayer.EnableVps {
		if cfg.Cloud.Properties.SoftLayer.VpsUseSsl {
			vps = vpsClient.New(httptransport.New(fmt.Sprintf("%s:%d", cfg.Cloud.Properties.SoftLayer.VpsHost, cfg.Cloud.Properties.SoftLayer.VpsPort),
				"v2", []string{"https"}), strfmt.Default).VM
		} else {

			vps = vpsClient.New(httptransport.New(fmt.Sprintf("%s:%d", cfg.Cloud.Properties.SoftLayer.VpsHost, cfg.Cloud.Properties.SoftLayer.VpsPort),
				"v2", []string{"http"}), strfmt.Default).VM
		}

	}

	repClientFactory := client.NewClientFactory(client.NewSoftLayerClientManager(softLayerClient, vps))
	client := repClientFactory.CreateClient()

	actionFactory := action.NewConcreteFactory(
		client,
		uuidGen,
		cfg,
		logger,
	)

	caller := dispatcher.NewJSONCaller()

	return dispatcher.NewJSON(actionFactory, caller, logger)
}
