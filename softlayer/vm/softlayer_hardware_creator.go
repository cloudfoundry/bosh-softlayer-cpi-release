package vm

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bmslc "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	sl "github.com/maximilien/softlayer-go/softlayer"

	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type baremetalCreator struct {
	softLayerClient        sl.Client
	bmsClient              bmslc.BmpClient
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
	vmFinder     Finder
}

func NewBaremetalCreator(vmFinder Finder, softLayerClient sl.Client, bmsClient bmslc.BmpClient, agentEnvServiceFactory AgentEnvServiceFactory, agentOptions AgentOptions, logger boshlog.Logger) VMCreator {
	bslcommon.TIMEOUT = 15 * time.Minute
	bslcommon.POLLING_INTERVAL = 5 * time.Second

	return &baremetalCreator{
		vmFinder:               vmFinder,
		softLayerClient:        softLayerClient,
		bmsClient:              bmsClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
	}
}

func (c *baremetalCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	hardwareId, err := c.createBaremetal(cloudProps.VmNamePrefix, cloudProps.BaremetalStemcell, cloudProps.BaremetalNetbootImage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Create baremetal error")
	}

	hardware, found, err := c.vmFinder.Find(hardwareId)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find hardware with id: %d.", hardwareId)
	}

	softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), c.logger)
	agentEnvService := c.agentEnvServiceFactory.New(softlayerFileService)

	// Update mbus url setting
	mbus, err := ParseMbusURL(c.agentOptions.Mbus, cloudProps.BoshIp)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
	}
	c.agentOptions.Mbus = mbus
	// Update blobstore setting
	switch c.agentOptions.Blobstore.Provider {
	case BlobstoreTypeDav:
		davConf := DavConfig(c.agentOptions.Blobstore.Options)
		UpdateDavConfig(&davConf, cloudProps.BoshIp)
	}

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot agent env for virtual guest with id: %d.", hardwareId)
	}

	err = agentEnvService.Update(agentEnv)
	if err != nil {
		return nil, bosherr.WrapError(err, "Updating VM's agent env")
	}

	if len(c.agentOptions.VcapPassword) > 0 {
		err = hardware.SetVcapPassword(c.agentOptions.VcapPassword)
		if err != nil {
			return nil, bosherr.WrapError(err, "Updating VM's vcap password")
		}
	}

	return hardware, nil
}

// Private methods
func (c *baremetalCreator) createBaremetal(server_name string, stemcell string, netboot_image string) (int, error) {
	provisioningBaremetalInfo := bmslc.ProvisioningBaremetalInfo{
		VmNamePrefix:     server_name,
		Bm_stemcell:      stemcell,
		Bm_netboot_image: netboot_image,
	}
	createBaremetalResponse, err := c.bmsClient.ProvisioningBaremetal(provisioningBaremetalInfo)
	if err != nil || createBaremetalResponse.Status != 200 || createBaremetalResponse.Data.TaskId == 0 {
		return 0, bosherr.WrapErrorf(err, "Failed to provisioning baremetal")
	}

	task_id := createBaremetalResponse.Data.TaskId
	bslcommon.TIMEOUT = 120 * time.Minute
	totalTime := time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {

		taskOutput, err := c.bmsClient.TaskJsonOutput(task_id, "task")
		if err != nil {
			return 0, bosherr.WrapErrorf(err, "Failed to get state with task_id: %d", task_id)
		}

		info := taskOutput.Data["info"].(map[string]interface{})
		switch info["status"].(string) {
		case "failed":
			return 0, bosherr.Errorf("Failed to install the stemcell: %v", taskOutput)

		case "completed":
			serverOutput, err := c.bmsClient.TaskJsonOutput(task_id, "server")
			if err != nil {
				return 0, bosherr.WrapErrorf(err, "Failed to get server_id with task_id: %d", task_id)
			}
			info = serverOutput.Data["info"].(map[string]interface{})
			return int(info["id"].(float64)), nil
		default:
			totalTime += bslcommon.POLLING_INTERVAL
			time.Sleep(bslcommon.POLLING_INTERVAL)
		}
	}

	return 0, bosherr.Error("Provisioning baremetal timeout")
}
