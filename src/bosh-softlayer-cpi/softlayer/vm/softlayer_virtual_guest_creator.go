package vm

import (
	"fmt"
	"net"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bsl "bosh-softlayer-cpi/softlayer/client"
	. "bosh-softlayer-cpi/softlayer/common"
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"
	datatypes "github.com/softlayer/softlayer-go/datatypes"

	"bosh-softlayer-cpi/util"

	snet "bosh-softlayer-cpi/softlayer/networks"
)

type softLayerVirtualGuestCreator struct {
	softLayerClient bsl.Client
	vmFinder        VMFinder
	agentOptions    AgentOptions
	registryOptions RegistryOptions
	featureOptions  FeatureOptions
	logger          boshlog.Logger
}

func NewSoftLayerCreator(vmFinder VMFinder, softLayerClient bsl.Client, agentOptions AgentOptions, featureOptions FeatureOptions, registryOptions RegistryOptions, logger boshlog.Logger) VMCreator {
	return &softLayerVirtualGuestCreator{
		vmFinder:        vmFinder,
		softLayerClient: softLayerClient,
		agentOptions:    agentOptions,
		registryOptions: registryOptions,
		featureOptions:  featureOptions,
		logger:          logger,
	}
}

func (sc *softLayerVirtualGuestCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks snet.Networks, env Environment) (VM, error) {
	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if sc.featureOptions.DisableOsReload {
				return sc.createBySoftlayer(agentID, stemcell, cloudProps, networks, env)
			} else {
				if len(network.IP) == 0 {
					return sc.createBySoftlayer(agentID, stemcell, cloudProps, networks, env)
				} else {
					return sc.createByOSReload(agentID, stemcell, cloudProps, networks, env)
				}

			}
		case "vip":
			return nil, bosherr.Error("SoftLayer CPI Not Support VIP netowrk")
		default:
			continue
		}
	}
	return nil, bosherr.Error("virtual guests must have exactly one dynamic network")
}

func (sc *softLayerVirtualGuestCreator) GetAgentOptions() AgentOptions { return sc.agentOptions }

// Private methods
func (sc *softLayerVirtualGuestCreator) createBySoftlayer(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks snet.Networks, env Environment) (VM, error) {
	userDataTypeContents, err := CreateUserDataForInstance(agentID, sc.registryOptions)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating virtualGuest userData")
	}

	publicVlanId, privateVlanId, err := GetVlanIds(sc.softLayerClient, networks)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting vlan ids from networks settings")
	}

	virtualGuestTemplate := CreateVirtualGuestTemplate(stemcell.Uuid(), cloudProps, userDataTypeContents, publicVlanId, privateVlanId)

	virtualGuest, err := sc.softLayerClient.CreateInstance(virtualGuestTemplate)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating virtualGuest")
	}

	if cloudProps.EphemeralDiskSize > 0 {
		sc.softLayerClient.AttachSecondDiskToInstance(*virtualGuest.Id, cloudProps.EphemeralDiskSize)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching second disk to virtualGuest with id `%d`", *virtualGuest.Id))
		}
	}

	vm, err := sc.vmFinder.Find(*virtualGuest.Id)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM with id `%d`", *virtualGuest.Id)
	}

	if cloudProps.DeployedByBoshCLI {
		err := UpdateEtcHostsOfBoshInit("/etc/hosts", fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		boshIP, err := GetLocalIPAddressOfGivenInterface("eth0")
		if err != nil {
			return nil, bosherr.WrapError(err, "Failed to get IP address of eth0 in local")
		}

		mbus, err := ParseMbusURL(sc.agentOptions.Mbus, boshIP)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		sc.agentOptions.Mbus = mbus

		switch sc.agentOptions.Blobstore.Provider {
		case BlobstoreTypeDav:
			davConf := DavConfig(sc.agentOptions.Blobstore.Options)
			UpdateDavConfig(&davConf, boshIP)
		}
	}

	networks, err = vm.ConfigureNetworks(networks)
	if err != nil {
		return nil, bosherr.WrapError(err, "Configuring VM's networking")
	}
	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, sc.agentOptions)
	err = vm.UpdateAgentEnv(agentEnv)
	if err != nil {
		return nil, bosherr.WrapError(err, "Updating VM's agent env")
	}

	if len(sc.agentOptions.VcapPassword) > 0 {
		err = vm.SetVcapPassword(sc.agentOptions.VcapPassword)
		if err != nil {
			return nil, bosherr.WrapError(err, "Updating VM's vcap password")
		}
	}

	return vm, nil
}

func (sc *softLayerVirtualGuestCreator) createByOSReload(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks snet.Networks, env Environment) (VM, error) {
	var virtualGuest datatypes.Virtual_Guest
	var err error

	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if util.IsPrivateSubnet(net.ParseIP(network.IP)) {
				virtualGuest, err = sc.softLayerClient.GetInstanceByPrimaryBackendIpAddress(network.IP)
			} else {
				virtualGuest, err = sc.softLayerClient.GetInstanceByPrimaryIpAddress(network.IP)
			}
			if err != nil || virtualGuest.Id == nil {
				return nil, bosherr.WrapErrorf(err, "Failed to find virtualGuest with ip address: %s", network.IP)
			}
		case "manual", "":
			continue
		default:
			return nil, bosherr.Errorf("unexpected network type: %s", network.Type)
		}
	}

	sc.logger.Info(SOFTLAYER_VM_CREATOR_LOG_TAG, fmt.Sprintf("OS reload on VirtualGuest %d using stemcell %d", virtualGuest.Id, stemcell.ID()))

	vm, err := sc.vmFinder.Find(*virtualGuest.Id)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding virtualGuest with id: %d", virtualGuest.Id)
	}

	err = vm.ReloadOS(stemcell)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reloading OS")
	}

	err = UpdateDeviceName(*virtualGuest.Id, sc.softLayerClient, cloudProps)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Updating device name after os_reload")
	}

	if cloudProps.EphemeralDiskSize > 0 {
		sc.softLayerClient.AttachSecondDiskToInstance(*virtualGuest.Id, cloudProps.EphemeralDiskSize)
		if err != nil {
			return nil, bosherr.WrapError(err, fmt.Sprintf("Attaching second disk to virtualGuest with id `%d`", *virtualGuest.Id))
		}
	}

	if cloudProps.DeployedByBoshCLI {
		err := UpdateEtcHostsOfBoshInit("/etc/hosts", fmt.Sprintf("%s  %s", vm.GetPrimaryBackendIP(), vm.GetFullyQualifiedDomainName()))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		boshIP, err := GetLocalIPAddressOfGivenInterface("eth0")
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Failed to get IP address of eth0 in local")
		}

		mbus, err := ParseMbusURL(sc.agentOptions.Mbus, boshIP)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		sc.agentOptions.Mbus = mbus

		switch sc.agentOptions.Blobstore.Provider {
		case BlobstoreTypeDav:
			davConf := DavConfig(sc.agentOptions.Blobstore.Options)
			UpdateDavConfig(&davConf, boshIP)
		}
	}

	vm, err = sc.vmFinder.Find(*virtualGuest.Id)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "refresh VM with id: %d after os_reload", virtualGuest.Id)
	}

	networks, err = vm.ConfigureNetworks(networks)
	if err != nil {
		return nil, bosherr.WrapError(err, "Configuring VM's networking")
	}
	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, sc.agentOptions)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Cannot create agent env for virtual guest with id: %d", vm.ID())
	}
	err = vm.UpdateAgentEnv(agentEnv)
	if err != nil {
		return nil, bosherr.WrapError(err, "Updating VM's agent env")
	}

	if len(sc.agentOptions.VcapPassword) > 0 {
		err = vm.SetVcapPassword(sc.agentOptions.VcapPassword)
		if err != nil {
			return nil, bosherr.WrapError(err, "Updating VM's vcap password")
		}
	}

	return vm, nil
}
