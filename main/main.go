package main

import (
	"flag"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	slclient "github.com/maximilien/softlayer-go/client"

	bslcaction "github.com/maximilien/bosh-softlayer-cpi/action"
	bslcdisp "github.com/maximilien/bosh-softlayer-cpi/api/dispatcher"
	bslctrans "github.com/maximilien/bosh-softlayer-cpi/api/transport"
)

const mainLogTag = "main"

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
	cpiVersion    = flag.Bool("version", false, "The version of CPI release")
)

func main() {
	logger, fs, cmdRunner, uuidGenerator := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	if *cpiVersion {
		os.Exit(0)
	}

	config, err := NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error(mainLogTag, "Loading config %s", err.Error())
		os.Exit(1)
	}

	dispatcher := buildDispatcher(config, logger, fs, cmdRunner, uuidGenerator)

	cli := bslctrans.NewCLI(os.Stdin, os.Stdout, dispatcher, logger)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error(mainLogTag, "Serving once %s", err)
		os.Exit(1)
	}
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem, boshsys.CmdRunner, boshuuid.Generator) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr, os.Stderr)

	fs := boshsys.NewOsFileSystem(logger)

	uuidGenerator := boshuuid.NewGenerator()

	cmdRunner := boshsys.NewExecCmdRunner(logger)

	return logger, fs, cmdRunner, uuidGenerator
}

func buildDispatcher(config Config, logger boshlog.Logger, fs boshsys.FileSystem, cmdRunner boshsys.CmdRunner, uuidGenerator boshuuid.Generator) bslcdisp.Dispatcher {
	softLayerClient := slclient.NewSoftLayerClient(config.SoftLayer.Username, config.SoftLayer.ApiKey)

	actionFactory := bslcaction.NewConcreteFactory(
		softLayerClient,
		config.Actions,
		logger,
		uuidGenerator,
		fs,
	)

	caller := bslcdisp.NewJSONCaller()

	return bslcdisp.NewJSON(actionFactory, caller, logger)
}
