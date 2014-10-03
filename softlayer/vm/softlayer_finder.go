package vm

import (
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"
)

const softLayerFinderLogTag = "SoftLayerFinder"

type SoftLayerFinder struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	logger boshlog.Logger
}

func NewSoftLayerFinder(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,

		logger: logger,
	}
}

func (f SoftLayerFinder) Find(vmID int) (VM, bool, error) {
	//DEBUG
	fmt.Println("SoftLayerFinder.Find")
	fmt.Printf("----> vmID: %#v\n", vmID)
	fmt.Println()
	os.Exit(0)
	//DEBUG

	f.logger.Debug(softLayerFinderLogTag, "Finding container with ID '%s'", vmID)

	//Find VM here using SL client

	f.logger.Debug(softLayerFinderLogTag, "Did not find container with ID '%s'", vmID)

	return nil, false, nil
}
