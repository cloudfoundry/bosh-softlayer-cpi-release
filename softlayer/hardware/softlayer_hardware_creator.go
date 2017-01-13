package hardware

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bmslc "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

type baremetalCreator struct {
	softLayerClient        sl.Client
	bmsClient              bmslc.BmpClient
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
	vmFinder     VMFinder
}

func NewBaremetalCreator(vmFinder VMFinder, softLayerClient sl.Client, bmsClient bmslc.BmpClient, agentOptions AgentOptions, logger boshlog.Logger) VMCreator {
	slh.TIMEOUT = 15 * time.Minute
	slh.POLLING_INTERVAL = 5 * time.Second

	return &baremetalCreator{
		vmFinder:        vmFinder,
		softLayerClient: softLayerClient,
		bmsClient:       bmsClient,
		agentOptions:    agentOptions,
		logger:          logger,
	}
}

func (c *baremetalCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if len(network.IP) == 0 {
				return c.createByBaremetal(agentID, stemcell, cloudProps, networks, env)
			} else {
				return c.createByOSReload(agentID, stemcell, cloudProps, networks, env)
			}
		case "manual":
			return nil, bosherr.Error("Manual networking is not currently supported")
		case "vip":
			return nil, bosherr.Error("SoftLayer Not Support VIP netowrk")
		default:
			return nil, bosherr.Errorf("Softlayer Not Support This Kind Of Network: %s", network.Type)
		}
	}

	return nil, nil
}

func (c *baremetalCreator) createByBaremetal(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	hardwareId, err := c.provisionBaremetal(cloudProps.VmNamePrefix, cloudProps.BaremetalStemcell, cloudProps.BaremetalNetbootImage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Create baremetal error")
	}

	hardware, found, err := c.vmFinder.Find(hardwareId)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find hardware with id: %d.", hardwareId)
	}

	var boshIP string
	if cloudProps.BoshIp != "" {
		boshIP = cloudProps.BoshIp
	} else {
		boshIP, err = GetLocalIPAddressOfGivenInterface(slh.NetworkInterface)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", slh.NetworkInterface))
		}
	}

	mbus, err := ParseMbusURL(c.agentOptions.Mbus, boshIP)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
	}
	c.agentOptions.Mbus = mbus

	switch c.agentOptions.Blobstore.Provider {
	case BlobstoreTypeDav:
		davConf := DavConfig(c.agentOptions.Blobstore.Options)
		UpdateDavConfig(&davConf, boshIP)
	}

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)

	err = hardware.UpdateAgentEnv(agentEnv)
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

func (c *baremetalCreator) createByOSReload(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	if len(cloudProps.BaremetalStemcell) == 0 {
		return nil, bosherr.Error("No stemcell provided to do os_reload.")
	}

	hardwareService, err := c.softLayerClient.GetSoftLayer_Hardware_Service()
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating HardwareService from SoftLayer client")
	}

	hardware, err := hardwareService.FindByIpAddress(networks.First().IP)
	if err != nil || hardware.Id == 0 {
		return nil, bosherr.WrapErrorf(err, "Could not find hardware by ip address: %s", networks.First().IP)
	}

	c.logger.Info(SOFTLAYER_VM_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on Hardware %d using stemcell %d", hardware.Id, stemcell.ID()))

	vm, found, err := c.vmFinder.Find(hardware.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find hardware with id: %d", hardware.Id)
	}

	err = vm.ReloadOSForBaremetal(cloudProps.BaremetalStemcell, cloudProps.BaremetalNetbootImage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to reload OS")
	}

	vm, found, err = c.vmFinder.Find(hardware.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find hardware with id: %d.", vm.ID())
	}

	// Update mbus url setting
	var boshIP string
	if cloudProps.BoshIp != "" {
		boshIP = cloudProps.BoshIp
	} else {
		boshIP, err = GetLocalIPAddressOfGivenInterface(slh.NetworkInterface)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", slh.NetworkInterface))
		}
	}

	mbus, err := ParseMbusURL(c.agentOptions.Mbus, boshIP)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
	}
	c.agentOptions.Mbus = mbus
	// Update blobstore setting
	switch c.agentOptions.Blobstore.Provider {
	case BlobstoreTypeDav:
		davConf := DavConfig(c.agentOptions.Blobstore.Options)
		UpdateDavConfig(&davConf, boshIP)
	}

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot create agent env for baremetal with id: %d.", vm.ID())
	}

	err = vm.UpdateAgentEnv(agentEnv)
	if err != nil {
		return nil, bosherr.WrapError(err, "Updating VM's agent env")
	}

	if len(c.agentOptions.VcapPassword) > 0 {
		err = vm.SetVcapPassword(c.agentOptions.VcapPassword)
		if err != nil {
			return nil, bosherr.WrapError(err, "Updating VM's vcap password")
		}
	}

	return vm, nil
}

// Private methods
func (c *baremetalCreator) provisionBaremetal(server_name string, stemcell string, netboot_image string) (int, error) {
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
	slh.TIMEOUT = 120 * time.Minute
	totalTime := time.Duration(0)
	for totalTime < slh.TIMEOUT {

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
			totalTime += slh.POLLING_INTERVAL
			time.Sleep(slh.POLLING_INTERVAL)
		}
	}

	return 0, bosherr.Error("Provisioning baremetal timeout")
}
