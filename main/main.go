package main

import (
	"flag"
	"os"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	
	slclient "github.com/maximilien/softlayer-go/client"

	bslcaction "github.com/maximilien/bosh-softlayer-cpi/action"
	bslcdisp "github.com/maximilien/bosh-softlayer-cpi/api/dispatcher"
	bslctrans "github.com/maximilien/bosh-softlayer-cpi/api/transport"
)

const mainLogTag = "main"

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
)

func main() {
	logger, fs, cmdRunner := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	config, err := NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error(mainLogTag, "Loading config %s", err.Error())
		os.Exit(1)
	}

	dispatcher := buildDispatcher(config, logger, fs, cmdRunner)

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

func buildDispatcher(
	config Config,
	logger boshlog.Logger,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
) bslcdisp.Dispatcher {

	softLayerClient := slclient.NewSoftLayerClient(config.SoftLayer.Username, config.SoftLayer.ApiKey)

	compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)

	actionFactory := bslcaction.NewConcreteFactory(
		softLayerClient,
		fs,
		cmdRunner,
		compressor,
		config.Actions,
		logger,
	)

	caller := bslcdisp.NewJSONCaller()

	return bslcdisp.NewJSON(actionFactory, caller, logger)
}
