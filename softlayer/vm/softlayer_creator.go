package vm

import (
	"fmt"

	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	common "github.com/maximilien/bosh-softlayer-cpi/common"
	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvmpool "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/pool"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

type SoftLayerCreator struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger

	OsReloadTimeout time.Duration
}

func NewSoftLayerCreator(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, agentOptions AgentOptions, logger boshlog.Logger) SoftLayerCreator {
	bslcommon.TIMEOUT = 20 * time.Minute
	bslcommon.POLLING_INTERVAL = 20 * time.Second
	bslcommon.MAX_RETRY_COUNT = 5

	return SoftLayerCreator{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
		OsReloadTimeout:        30 * time.Second,
	}
}

func (c SoftLayerCreator) CreateNewVM(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
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
	vm := NewSoftLayerVM(virtualGuest.Id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger, TIMEOUT_TRANSACTIONS_CREATE_VM)

	if strings.ToUpper(common.GetOSEnvVariable("OS_RELOAD_ENABLED", "TRUE")) == "TRUE" {
		db, err := bslcvmpool.OpenDB()
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, "Opening DB")
		}

		vmInfoDB := bslcvmpool.NewVMInfoDB(vm.id, virtualGuestTemplate.Hostname+"."+virtualGuestTemplate.Domain, "t", stemcell.Uuid(), agentID, c.logger, db)
		err = vmInfoDB.InsertVMInfo()
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, "Failed to insert the record into VM pool DB")
		}
	}

	return vm, nil
}


func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	if strings.ToUpper(common.GetOSEnvVariable("OS_RELOAD_ENABLED", "TRUE")) == "FALSE" {
		return c.CreateNewVM(agentID, stemcell, cloudProps, networks, env)
	}

	if strings.Contains(cloudProps.VmNamePrefix, "-worker") {
		vm, err := c.CreateNewVM(agentID, stemcell, cloudProps, networks, env)
		return vm, err
	}

	err := bslcvmpool.InitVMPoolDB()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Failed to initialize VM pool DB")
	}

	db, err := bslcvmpool.OpenDB()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Opening DB")
	}

	vmInfoDB := bslcvmpool.NewVMInfoDB(0, "", "f", "", agentID, c.logger, db)
	defer vmInfoDB.CloseDB()

	err = vmInfoDB.QueryVMInfobyAgentID()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Failed to query VM info by given agent ID "+agentID)
	}

	if vmInfoDB.VmProperties.Id != 0 {
		c.logger.Info(softLayerCreatorLogTag, fmt.Sprintf("OS reload on the server id %d with stemcell %d", vmInfoDB.VmProperties.Id, stemcell.ID()))

		agentEnvService := c.agentEnvServiceFactory.New(vmInfoDB.VmProperties.Id)
		vm := NewSoftLayerVM(vmInfoDB.VmProperties.Id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger, TIMEOUT_TRANSACTIONS_OSRELOAD_VM)

		vm.ReloadOS(stemcell, c.OsReloadTimeout)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, "Failed to reload OS")
		}

		c.logger.Info(softLayerCreatorLogTag, fmt.Sprintf("Updated in_use flag to 't' for the VM %d in VM pool", vmInfoDB.VmProperties.Id))
		vmInfoDB.VmProperties.InUse = "t"
		err = vmInfoDB.UpdateVMInfoByID()
		if err != nil {
			return vm, bosherr.WrapError(err, fmt.Sprintf("Failed to query VM info by given ID %d", vm.ID()))
		} else {
			return vm, nil
		}
	}

	vmInfoDB.VmProperties.InUse = ""
	err = vmInfoDB.QueryVMInfobyAgentID()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Failed to query VM info by given agent ID "+agentID)
	}

	if vmInfoDB.VmProperties.Id != 0 {
		return SoftLayerVM{}, bosherr.WrapError(err, "Wrong in_use status in VM with agent ID "+agentID+", Do not create a new VM")
	} else {
		return c.CreateNewVM(agentID, stemcell, cloudProps, networks, env)
	}

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
