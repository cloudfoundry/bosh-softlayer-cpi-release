package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bsl "bosh-softlayer-cpi/softlayer/client"
	. "bosh-softlayer-cpi/softlayer/common"

	"bosh-softlayer-cpi/util"
)

type softLayerFinder struct {
	softLayerClient        bsl.Client
	agentEnvServiceFactory AgentEnvServiceFactory
	logger                 boshlog.Logger
}

func NewSoftLayerFinder(softLayerClient bsl.Client, agentEnvServiceFactory AgentEnvServiceFactory, logger boshlog.Logger) VMFinder {
	return &softLayerFinder{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		logger:                 logger,
	}
}

func (sf *softLayerFinder) Find(cid int) (VM, error) {
	virtualGuest, err := sf.softLayerClient.GetInstance(cid, bsl.INSTANCE_DETAIL_MASK)
	if err != nil {
		return nil, bosherr.Errorf("Getting instance with id `%d`", cid)
	}

	vm := NewSoftLayerVirtualGuest(&virtualGuest, sf.softLayerClient, util.GetSshClient(), sf.logger)
	agentEnvService := sf.agentEnvServiceFactory.New(vm, NewSoftlayerFileService(util.GetSshClient(), sf.logger))
	vm.SetAgentEnvService(agentEnvService)
	return vm, nil
}
