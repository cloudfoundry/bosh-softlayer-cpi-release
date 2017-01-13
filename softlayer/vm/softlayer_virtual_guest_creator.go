package vm

import (
	"fmt"
	"net"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"

	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type softLayerVirtualGuestCreator struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
	vmFinder     VMFinder

	featureOptions FeatureOptions
}

func NewSoftLayerCreator(vmFinder VMFinder, softLayerClient sl.Client, agentOptions AgentOptions, logger boshlog.Logger, featureOptions FeatureOptions) VMCreator {
	slhelper.TIMEOUT = 120 * time.Minute
	slhelper.POLLING_INTERVAL = 5 * time.Second

	return &softLayerVirtualGuestCreator{
		vmFinder:        vmFinder,
		softLayerClient: softLayerClient,
		agentOptions:    agentOptions,
		logger:          logger,
		featureOptions:  featureOptions,
	}
}

func (c *softLayerVirtualGuestCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if cloudProps.DisableOsReload || c.featureOptions.DisableOsReload {
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

// Private methods
func (c *softLayerVirtualGuestCreator) createBySoftlayer(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
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

	slhelper.TIMEOUT = 4 * time.Hour
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
