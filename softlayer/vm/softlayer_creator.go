package vm

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

type SoftLayerCreator struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
}

func NewSoftLayerCreator(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, agentOptions AgentOptions, logger boshlog.Logger) SoftLayerCreator {
	bslcommon.TIMEOUT = 60 * time.Minute
	bslcommon.POLLING_INTERVAL = 20 * time.Second
	bslcommon.MAX_RETRY_COUNT = 5

	return SoftLayerCreator{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
	}
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestTemplate, err := CreateVirtualGuestTemplate(agentID, stemcell, cloudProps, networks, env, c.agentOptions)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating virtual guest template")
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(virtualGuestTemplate)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuest from SoftLayer client")
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = bslcommon.WaitForVirtualGuest(c.softLayerClient, virtualGuest.Id, "RUNNING", bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id `%d`", virtualGuest.Id))
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
		}
	}

	agentEnvService := c.agentEnvServiceFactory.New(virtualGuest.Id)

	vm := NewSoftLayerVM(virtualGuest.Id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

	return vm, nil
}

// Private methods

func (c SoftLayerCreator) resolveNetworkIP(networks Networks) (string, error) {
	var network Network

	switch len(networks) {
	case 0:
		return "", bosherr.Error("Expected exactly one network; received zero")
	case 1:
		network = networks.First()
	default:
		return "", bosherr.Error("Expected exactly one network; received multiple")
	}

	if network.IsDynamic() {
		return "", nil
	}

	return network.IP, nil
}
