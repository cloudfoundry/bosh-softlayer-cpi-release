package pool

import (
	"fmt"
	"time"
	"net"

	strfmt "github.com/go-openapi/strfmt"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	datatypes "github.com/maximilien/softlayer-go/data_types"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	sl "github.com/maximilien/softlayer-go/softlayer"
	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client"
	operations "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/models"
	"github.com/cloudfoundry/bosh-softlayer-cpi/common"
)

const SOFTLAYER_POOL_CREATOR_LOG_TAG = "SoftLayerPoolCreator"

type softLayerPoolCreator struct {
	softLayerClient        sl.Client
	softLayerVmPoolClient  *client.SoftLayerVMPool

	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions           AgentOptions

	logger                 boshlog.Logger

	vmFinder               Finder

	featureOptions         FeatureOptions
}

func NewSoftLayerPoolCreator(vmFinder Finder, softLayerVmPoolClient *client.SoftLayerVMPool, softLayerClient sl.Client, agentOptions AgentOptions, logger boshlog.Logger, featureOptions FeatureOptions) VMCreator {
	bslcommon.TIMEOUT = 120 * time.Minute
	bslcommon.POLLING_INTERVAL = 5 * time.Second

	return &softLayerPoolCreator{
		softLayerVmPoolClient: softLayerVmPoolClient,
		softLayerClient: softLayerClient,
		agentOptions:    agentOptions,
		logger:          logger,
		vmFinder:        vmFinder,
		featureOptions:  featureOptions,
	}
}

func (c *softLayerPoolCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	for _, network := range networks {
		switch network.Type {
			case "dynamic":
				if len(network.IP) == 0 {
					return c.createFromVMPool(agentID, stemcell, cloudProps, networks, env)
				} else {
					return c.createByOSReload(agentID, stemcell, cloudProps, networks, env)
				}
			case "vip":
				return nil, bosherr.Error("SoftLayer Not Support VIP netowrk")
			default:
				continue
		}
        }
        return nil, bosherr.Error("virtual guests must have exactly one dynamic network")
}

// Private methods
func (c *softLayerPoolCreator) createFromVMPool(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	var err error
	virtualGuestTemplate, err := CreateVirtualGuestTemplate(stemcell, cloudProps, networks)
	filter := &models.VMFilter{
		CPU:         int32(virtualGuestTemplate.StartCpus),
		MemoryMb:    int32(virtualGuestTemplate.MaxMemory),
		PrivateVlan: int32(virtualGuestTemplate.PrimaryBackendNetworkComponent.NetworkVlan.Id),
		PublicVlan:  int32(virtualGuestTemplate.PrimaryNetworkComponent.NetworkVlan.Id),
		State: models.StateFree,
	}
	findVMsResp, err := c.softLayerVmPoolClient.VM.FindVmsByFilters(operations.NewFindVmsByFiltersParams().WithBody(filter))
	if err != nil {
		return nil, bosherr.WrapError(err, "Finding vms from pool")
	}

	if len (findVMsResp.Payload.Vms) == 0 {
	        sl_vm, err := c.createBySoftlayer(agentID, stemcell, cloudProps, networks, env)
		if err != nil {
			return nil, bosherr.WrapError(err, "Creating vm in SoftLayer")
		}

		slPoolVm := &models.VM{
			Cid: int32(sl_vm.ID()),
			CPU: int32(virtualGuestTemplate.StartCpus),
			MemoryMb: int32(virtualGuestTemplate.MaxMemory),
			IP:  strfmt.IPv4(sl_vm.GetPrimaryBackendIP()),
			Hostname: sl_vm.GetFullyQualifiedDomainName(),
			PrivateVlan: int32(virtualGuestTemplate.PrimaryBackendNetworkComponent.NetworkVlan.Id),
			PublicVlan: int32(virtualGuestTemplate.PrimaryNetworkComponent.NetworkVlan.Id),
			State: models.StateUsing,
		}
		_, err = c.softLayerVmPoolClient.VM.AddVM(operations.NewAddVMParams().WithBody(slPoolVm))
		if err != nil {
			return nil, bosherr.WrapError(err, "Adding vms from pool")
		}
		c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("Added vm %d to pool successfully", sl_vm.ID()))

		return sl_vm, nil
	} else {
		var vm *models.VM
		var virtualGuestId int

		vm = findVMsResp.Payload.Vms[0]
		virtualGuestId = int(vm.Cid)

		provision := &models.VMState{
			State: models.StateProvisioning,
		}

		c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("Picking up the first vm %d from result set using stemcell %d to do os reload", vm.Cid, stemcell.ID()))

		_, err = c.softLayerVmPoolClient.VM.UpdateVMWithState(operations.NewUpdateVMWithStateParams().WithBody(provision).WithCid(vm.Cid))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating state of vm %d in pool to provisioning", virtualGuestId)
		}

		c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on VirtualGuest %d using stemcell %d", virtualGuestId, stemcell.ID()))

		sl_vm_os, err := c.oSReloadVMInPool(virtualGuestId, agentID, stemcell, cloudProps, networks, env)
		if err != nil {
			free := &models.VMState{
				State: models.StateFree,
			}
			_, err = c.softLayerVmPoolClient.VM.UpdateVMWithState(operations.NewUpdateVMWithStateParams().WithBody(free).WithCid(vm.Cid))
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Updating state of vm %d in pool to free", virtualGuestId)
			}

			return nil, bosherr.WrapError(err, "Os reloading vm in SoftLayer")
		}

		using := &models.VMState{
			State: models.StateUsing,
		}
		_, err = c.softLayerVmPoolClient.VM.UpdateVMWithState(operations.NewUpdateVMWithStateParams().WithBody(using).WithCid(vm.Cid))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating state of vm %d in pool to using", virtualGuestId)
		}

		c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("vm %d using stemcell %d os reload completed", vm.Cid, stemcell.ID()))

		return sl_vm_os, nil
	}
	return nil, nil
}

func (c *softLayerPoolCreator) createBySoftlayer(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestTemplate, err := CreateVirtualGuestTemplate(stemcell, cloudProps, networks)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VirtualGuest template")
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(virtualGuestTemplate)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VirtualGuest from SoftLayer client")
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = bslcommon.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, virtualGuest.Id, "Service Setup")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", virtualGuest.Id)
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
		}
	}

	vm, found, err := c.vmFinder.Find(virtualGuest.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find VirtualGuest with id: %d.", virtualGuest.Id)
	}

	if len(cloudProps.BoshIp) == 0 {
		UpdateEtcHostsOfBoshInit(fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		mbus, err := ParseMbusURL(c.agentOptions.Mbus, vm.GetPrimaryBackendIP())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus
	} else {
		mbus, err := ParseMbusURL(c.agentOptions.Mbus, cloudProps.BoshIp)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus

		switch c.agentOptions.Blobstore.Provider {
		case BlobstoreTypeDav:
			davConf := DavConfig(c.agentOptions.Blobstore.Options)
			UpdateDavConfig(&davConf, cloudProps.BoshIp)
		}
	}

	vm.ConfigureNetworks2(networks)

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)

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

func (c *softLayerPoolCreator) createByOSReload(agentID string , stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	var virtualGuest datatypes.SoftLayer_Virtual_Guest

	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if common.IsPrivateSubnet(net.ParseIP(network.IP)) {
				virtualGuest, err = virtualGuestService.GetObjectByPrimaryBackendIpAddress(network.IP)
			} else {
				virtualGuest, err = virtualGuestService.GetObjectByPrimaryIpAddress(network.IP)
			}
			if err != nil || virtualGuest.Id == 0 {
				return nil, bosherr.WrapErrorf(err, "Could not find VirtualGuest by ip address: %s", network.IP)
			}
		case "manual", "":
			continue
		default:
			return nil, bosherr.Errorf("unexpected network type: %s", network.Type)
		}
	}

	c.logger.Info(SOFTLAYER_VM_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on VirtualGuest %d using stemcell %d", virtualGuest.Id, stemcell.ID()))

	vm, found, err := c.vmFinder.Find(virtualGuest.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find virtualGuest with id: %d", virtualGuest.Id)
	}

	bslcommon.TIMEOUT = 4 * time.Hour
	err = vm.ReloadOS(stemcell)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to reload OS")
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = bslcommon.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, vm.ID(), "Service Setup")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", vm.ID())
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, vm.ID(), cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", vm.ID()))
		}
	}

	if len(cloudProps.BoshIp) == 0 {
		UpdateEtcHostsOfBoshInit(fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		mbus, err := ParseMbusURL(c.agentOptions.Mbus, vm.GetPrimaryBackendIP())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus
	} else {
		mbus, err := ParseMbusURL(c.agentOptions.Mbus, cloudProps.BoshIp)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus

		switch c.agentOptions.Blobstore.Provider {
		case BlobstoreTypeDav:
			davConf := DavConfig(c.agentOptions.Blobstore.Options)
			UpdateDavConfig(&davConf, cloudProps.BoshIp)
		}
	}

	vm, found, err = c.vmFinder.Find(virtualGuest.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "refresh VM with id: %d after os_reload", virtualGuest.Id)
	}

	vm.ConfigureNetworks2(networks)

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot create agent env for virtual guest with id: %d", vm.ID())
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

func (c *softLayerPoolCreator) oSReloadVMInPool (cid int, agentID string , stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	c.logger.Info(SOFTLAYER_VM_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on VirtualGuest %d using stemcell %d", cid, stemcell.ID()))

	vm, found, err := c.vmFinder.Find(cid)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find virtualGuest with id: %d", cid)
	}

	bslcommon.TIMEOUT = 4 * time.Hour
	err = vm.ReloadOS(stemcell)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to reload OS")
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = bslcommon.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, vm.ID(), "Service Setup")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", vm.ID())
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, vm.ID(), cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", vm.ID()))
		}
	}

	if len(cloudProps.BoshIp) == 0 {
		UpdateEtcHostsOfBoshInit(fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		mbus, err := ParseMbusURL(c.agentOptions.Mbus, vm.GetPrimaryBackendIP())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus
	} else {
		mbus, err := ParseMbusURL(c.agentOptions.Mbus, cloudProps.BoshIp)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus

		switch c.agentOptions.Blobstore.Provider {
		case BlobstoreTypeDav:
			davConf := DavConfig(c.agentOptions.Blobstore.Options)
			UpdateDavConfig(&davConf, cloudProps.BoshIp)
		}
	}

	vm, found, err = c.vmFinder.Find(cid)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "refresh VM with id: %d after os_reload", cid)
	}

	vm.ConfigureNetworks2(networks)

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot create agent env for virtual guest with id: %d", vm.ID())
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