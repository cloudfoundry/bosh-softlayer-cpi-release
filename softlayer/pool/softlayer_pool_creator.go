package pool

import (
	"fmt"
	"net"
	"time"

	"github.com/go-openapi/strfmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	datatypes "github.com/maximilien/softlayer-go/data_types"

	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	operations "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	"github.com/cloudfoundry/bosh-softlayer-cpi/util"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/models"
)

const SOFTLAYER_POOL_CREATOR_LOG_TAG = "SoftLayerPoolCreator"

type softLayerPoolCreator struct {
	softLayerClient       sl.Client
	softLayerVmPoolClient operations.SoftLayerPoolClient

	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions

	logger boshlog.Logger

	vmFinder VMFinder

	featureOptions FeatureOptions
}

func NewSoftLayerPoolCreator(vmFinder VMFinder, softLayerVmPoolClient operations.SoftLayerPoolClient, softLayerClient sl.Client, agentOptions AgentOptions, logger boshlog.Logger, featureOptions FeatureOptions) VMCreator {
	slhelper.TIMEOUT = 120 * time.Minute
	slhelper.POLLING_INTERVAL = 5 * time.Second

	return &softLayerPoolCreator{
		softLayerVmPoolClient: softLayerVmPoolClient,
		softLayerClient:       softLayerClient,
		agentOptions:          agentOptions,
		logger:                logger,
		vmFinder:              vmFinder,
		featureOptions:        featureOptions,
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
		State:       models.StateFree,
	}
	orderVmResp, err := c.softLayerVmPoolClient.OrderVMByFilter(operations.NewOrderVMByFilterParams().WithBody(filter))
	if err != nil {
		_, ok := err.(*operations.OrderVMByFilterNotFound)
		if !ok {
			return nil, bosherr.WrapError(err, "Ordering vm from pool")
		} else {
			sl_vm, err := c.createBySoftlayer(agentID, stemcell, cloudProps, networks, env)
			if err != nil {
				return nil, bosherr.WrapError(err, "Creating vm in SoftLayer")
			}
			slPoolVm := &models.VM{
				Cid:         int32(sl_vm.ID()),
				CPU:         int32(virtualGuestTemplate.StartCpus),
				MemoryMb:    int32(virtualGuestTemplate.MaxMemory),
				IP:          strfmt.IPv4(sl_vm.GetPrimaryBackendIP()),
				Hostname:    sl_vm.GetFullyQualifiedDomainName(),
				PrivateVlan: int32(virtualGuestTemplate.PrimaryBackendNetworkComponent.NetworkVlan.Id),
				PublicVlan:  int32(virtualGuestTemplate.PrimaryNetworkComponent.NetworkVlan.Id),
				State:       models.StateUsing,
			}
			_, err = c.softLayerVmPoolClient.AddVM(operations.NewAddVMParams().WithBody(slPoolVm))
			if err != nil {
				return nil, bosherr.WrapError(err, "Adding vm into pool")
			}
			c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("Added vm %d to pool successfully", sl_vm.ID()))

			return sl_vm, nil
		}
	}
	var vm *models.VM
	var virtualGuestId int

	vm = orderVmResp.Payload.VM
	virtualGuestId = int((*vm).Cid)

	c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on VirtualGuest %d using stemcell %d", virtualGuestId, stemcell.ID()))

	sl_vm_os, err := c.oSReloadVMInPool(virtualGuestId, agentID, stemcell, cloudProps, networks, env)
	if err != nil {
		return nil, bosherr.WrapError(err, "Os reloading vm in SoftLayer")
	}

	virtualGuest, err := slhelper.GetObjectDetailsOnVirtualGuest(c.softLayerClient, virtualGuestId)
	if err != nil {
		return nil, bosherr.WrapError(err, fmt.Sprintf("Getting virtual guest %d details from SoftLayer", virtualGuestId))
	}
	deviceName := &models.VM{
		Cid:         int32(virtualGuestId),
		CPU:         int32(virtualGuest.StartCpus),
		MemoryMb:    int32(virtualGuest.MaxMemory),
		IP:          strfmt.IPv4(virtualGuest.PrimaryBackendIpAddress),
		Hostname:    cloudProps.VmNamePrefix + "." + cloudProps.Domain,
		PrivateVlan: int32(virtualGuest.PrimaryBackendNetworkComponent.NetworkVlan.Id),
		PublicVlan:  int32(virtualGuest.PrimaryNetworkComponent.NetworkVlan.Id),
		State:       models.StateUsing,
	}
	_, err = c.softLayerVmPoolClient.UpdateVM(operations.NewUpdateVMParams().WithBody(deviceName))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Updating the hostname of vm %d in pool to using", virtualGuestId)
	}

	c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("vm %d using stemcell %d os reload completed", virtualGuestId, stemcell.ID()))

	return sl_vm_os, nil
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
		err = slhelper.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, virtualGuest.Id, "Service Setup")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", virtualGuest.Id)
		}
	} else {
		err = slhelper.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
		}
	}

	vm, found, err := c.vmFinder.Find(virtualGuest.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find VirtualGuest with id: %d.", virtualGuest.Id)
	}

	if cloudProps.NotDeployedByDirector {
		err := UpdateEtcHostsOfBoshInit(slhelper.LocalDNSConfigurationFile, fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}

		mbus, err := ParseMbusURL(c.agentOptions.Mbus, vm.GetPrimaryBackendIP())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus
	} else {
		var boshIP string
		if cloudProps.BoshIp != "" {
			boshIP = cloudProps.BoshIp
		} else {
			boshIP, err = GetLocalIPAddressOfGivenInterface(slhelper.NetworkInterface)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", slhelper.NetworkInterface))
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

func (c *softLayerPoolCreator) createByOSReload(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	var virtualGuest datatypes.SoftLayer_Virtual_Guest

	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if util.IsPrivateSubnet(net.ParseIP(network.IP)) {
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

	c.logger.Info(SOFTLAYER_POOL_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on VirtualGuest %d using stemcell %d", virtualGuest.Id, stemcell.ID()))

	vm, found, err := c.vmFinder.Find(virtualGuest.Id)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find virtualGuest with id: %d", virtualGuest.Id)
	}

	slhelper.TIMEOUT = 4 * time.Hour
	err = vm.ReloadOS(stemcell)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to reload OS")
	}

	err = UpdateDeviceName(vm.ID(), virtualGuestService, cloudProps)
	if err != nil {
		return nil, err
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = slhelper.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, vm.ID(), "Service Setup")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", vm.ID())
		}
	} else {
		err = slhelper.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, vm.ID(), cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", vm.ID()))
		}
	}

	if cloudProps.NotDeployedByDirector {
		err := UpdateEtcHostsOfBoshInit(slhelper.LocalDNSConfigurationFile, fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}

		mbus, err := ParseMbusURL(c.agentOptions.Mbus, vm.GetPrimaryBackendIP())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus
	} else {
		var boshIP string
		if cloudProps.BoshIp != "" {
			boshIP = cloudProps.BoshIp
		} else {
			boshIP, err = GetLocalIPAddressOfGivenInterface(slhelper.NetworkInterface)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", slhelper.NetworkInterface))
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

func (c *softLayerPoolCreator) oSReloadVMInPool(cid int, agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	vm, found, err := c.vmFinder.Find(cid)
	if err != nil || !found {
		return nil, bosherr.WrapErrorf(err, "Cannot find virtualGuest with id: %d", cid)
	}

	slhelper.TIMEOUT = 4 * time.Hour
	err = vm.ReloadOS(stemcell)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Failed to do os_reload against %d", cid)
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	err = UpdateDeviceName(cid, virtualGuestService, cloudProps)
	if err != nil {
		return nil, err
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = slhelper.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, vm.ID(), "Service Setup")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", vm.ID())
		}
	} else {
		err = slhelper.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, vm.ID(), cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", vm.ID()))
		}
	}

	if cloudProps.NotDeployedByDirector {
		err := UpdateEtcHostsOfBoshInit(slhelper.LocalDNSConfigurationFile, fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}

		mbus, err := ParseMbusURL(c.agentOptions.Mbus, vm.GetPrimaryBackendIP())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		c.agentOptions.Mbus = mbus
	} else {
		var boshIP string
		if cloudProps.BoshIp != "" {
			boshIP = cloudProps.BoshIp
		} else {
			boshIP, err = GetLocalIPAddressOfGivenInterface(slhelper.NetworkInterface)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", slhelper.NetworkInterface))
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
