package vm

import (
	"fmt"
	"net"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "bosh-softlayer-cpi/softlayer/common"
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/util"

	slhelper "bosh-softlayer-cpi/softlayer/common/helpers"
)

type softLayerVirtualGuestCreator struct {
	softLayerClient sl.Client
	vmFinder        VMFinder
	agentOptions    AgentOptions
	registryOptions RegistryOptions
	featureOptions  FeatureOptions
	logger          boshlog.Logger
}

func NewSoftLayerCreator(vmFinder VMFinder, softLayerClient sl.Client, agentOptions AgentOptions, featureOptions FeatureOptions, registryOptions RegistryOptions, logger boshlog.Logger) VMCreator {
	return &softLayerVirtualGuestCreator{
		vmFinder:        vmFinder,
		softLayerClient: softLayerClient,
		agentOptions:    agentOptions,
		registryOptions: registryOptions,
		featureOptions:  featureOptions,
		logger:          logger,
	}
}

func (c *softLayerVirtualGuestCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	api.TIMEOUT = 120 * time.Minute
	api.POLLING_INTERVAL = 5 * time.Second

	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if c.featureOptions.DisableOsReload {
				return c.createBySoftlayer(agentID, stemcell, cloudProps, networks, env)
			} else {
				if len(network.IP) == 0 {
					return c.createBySoftlayer(agentID, stemcell, cloudProps, networks, env)
				} else {
					return c.createByOSReload(agentID, stemcell, cloudProps, networks, env)
				}

			}
		case "vip":
			return nil, bosherr.Error("SoftLayer Not Support VIP netowrk")
		default:
			continue
		}
	}

	return nil, bosherr.Error("virtual guests must have exactly one dynamic network")
}

func (c *softLayerVirtualGuestCreator) GetAgentOptions() AgentOptions { return c.agentOptions }

// Private methods
func (c *softLayerVirtualGuestCreator) createBySoftlayer(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	userDataTypeContents, err := CreateUserDataForInstance(agentID, c.registryOptions)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VirtualGuest UserData")
	}

	publicVlanId, privateVlanId, err := slhelper.GetVlanIds(c.softLayerClient, networks)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting Vlan Ids from networks settings")
	}

	virtualGuestTemplate, err := CreateVirtualGuestTemplate(stemcell.Uuid(), cloudProps, userDataTypeContents, publicVlanId, privateVlanId)
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

	if cloudProps.DeployedByBoshCLI {
		err := UpdateEtcHostsOfBoshInit(api.LocalDNSConfigurationFile, fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		var boshIP string
		if cloudProps.BoshIp != "" {
			boshIP = cloudProps.BoshIp
		} else {
			boshIP, err = GetLocalIPAddressOfGivenInterface(api.NetworkInterface)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", api.NetworkInterface))
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

	networks, err = vm.ConfigureNetworks(networks)
	if err != nil {
		return nil, bosherr.WrapError(err, "Configuring VM's networking")
	}

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

func (c *softLayerVirtualGuestCreator) createByOSReload(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
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
			break
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

	api.TIMEOUT = 4 * time.Hour
	err = vm.ReloadOS(stemcell)
	if err != nil {
		return nil, bosherr.WrapError(err, "Failed to reload OS")
	}

	err = UpdateDeviceName(virtualGuest.Id, virtualGuestService, cloudProps)
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

	if cloudProps.DeployedByBoshCLI {
		err := UpdateEtcHostsOfBoshInit(api.LocalDNSConfigurationFile, fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		var boshIP string
		if cloudProps.BoshIp != "" {
			boshIP = cloudProps.BoshIp
		} else {
			boshIP, err = GetLocalIPAddressOfGivenInterface(api.NetworkInterface)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get IP address of %s in local", api.NetworkInterface))
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

	vm.ConfigureNetworks(networks)

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
