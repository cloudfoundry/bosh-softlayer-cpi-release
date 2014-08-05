package main

import (
	"flag"
	"os"

	boshlog "bosh/logger"
	boshcmd "bosh/platform/commands"
	boshsys "bosh/system"
	boshuuid "bosh/uuid"
	wrdnclient "github.com/cloudfoundry-incubator/garden/client"
	wrdnconn "github.com/cloudfoundry-incubator/garden/client/connection"

	bslcaction "github.com/maximilien/bosh-softlayer-cpi/action"
	bslcdisp "github.com/maximilien/bosh-softlayer-cpi/api/dispatcher"
	bslctrans "github.com/maximilien/bosh-softlayer-cpi/api/transport"
	bslcutil "github.com/maximilien/bosh-softlayer-cpi/util"
)

const mainLogTag = "main"

var (
	configPathOpt = flag.String("configPath", "", "Path to configuration file")
)

func main() {
	logger, fs, cmdRunner, uuidGen := basicDeps()

	defer logger.HandlePanic("Main")

	flag.Parse()

	config, err := NewConfigFromPath(*configPathOpt, fs)
	if err != nil {
		logger.Error(mainLogTag, "Loading config %s", err.Error())
		os.Exit(1)
	}

	dispatcher := buildDispatcher(config, logger, fs, cmdRunner, uuidGen)

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

	cmdRunner := boshsys.NewExecCmdRunner(logger)

	uuidGen := boshuuid.NewGenerator()

	return logger, fs, cmdRunner, uuidGen
}

func buildDispatcher(
	config Config,
	logger boshlog.Logger,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
) bslcdisp.Dispatcher {
	wardenConn := wrdnconn.New(
		config.Warden.ConnectNetwork,
		config.Warden.ConnectAddress,
	)

	wardenClient := wrdnclient.New(wardenConn)

	compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)

	sleeper := bslcutil.RealSleeper{}

	actionFactory := bslcaction.NewConcreteFactory(
		wardenClient,
		fs,
		cmdRunner,
		uuidGen,
		compressor,
		sleeper,
		config.Actions,
		logger,
	)

	caller := bslcdisp.NewJSONCaller()

	return bslcdisp.NewJSON(actionFactory, caller, logger)
}
