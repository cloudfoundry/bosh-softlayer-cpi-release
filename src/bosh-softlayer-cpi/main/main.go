package main

import (
	"flag"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bslcaction "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	bslcdisp "github.com/cloudfoundry/bosh-softlayer-cpi/api/dispatcher"
	bslctrans "github.com/cloudfoundry/bosh-softlayer-cpi/api/transport"

	"github.com/cloudfoundry/bosh-softlayer-cpi/config"
)

const mainLogTag = "main"

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
	cpiVersion    = flag.Bool("version", false, "The version of CPI release")
)

func main() {
	logger, fs, cmdRunner := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	if *cpiVersion {
		os.Exit(0)
	}

	config, err := config.NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error(mainLogTag, "Loading config %s", err.Error())
		os.Exit(1)
	}

	dispatcher := buildDispatcher(config, logger, cmdRunner)

	cli := bslctrans.NewCLI(os.Stdin, os.Stdout, dispatcher, logger)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error(mainLogTag, "Serving once %s", err)
		os.Exit(1)
	}
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem, boshsys.CmdRunner) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr, os.Stderr)

	fs := boshsys.NewOsFileSystem(logger)

	cmdRunner := boshsys.NewExecCmdRunner(logger)

	return logger, fs, cmdRunner
}

func buildDispatcher(config config.Config, logger boshlog.Logger, cmdRunner boshsys.CmdRunner) bslcdisp.Dispatcher {
	actionFactory := bslcaction.NewConcreteFactory(
		config.Cloud.Properties,
		logger,
	)

	caller := bslcdisp.NewJSONCaller()

	return bslcdisp.NewJSON(actionFactory, caller, logger)
}
