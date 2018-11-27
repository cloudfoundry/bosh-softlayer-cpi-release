package action

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/bluebosh/goodhosts"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"
	boslc "bosh-softlayer-cpi/softlayer/client"
	boslconfig "bosh-softlayer-cpi/softlayer/config"
	"bosh-softlayer-cpi/softlayer/stemcell_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

type CreateVM struct {
	stemcellService     stemcell.Service
	virtualGuestService instance.Service
	registryClient      registry.Client
	registryOptions     registry.ClientOptions
	agentOptions        registry.AgentOptions
	softlayerOptions    boslconfig.Config
	localDNSConfigFile  string
}

func NewCreateVM(
	stemcellService stemcell.Service,
	virtualGuestService instance.Service,
	registryClient registry.Client,
	registryOptions registry.ClientOptions,
	agentOptions registry.AgentOptions,
	softlayerOptions boslconfig.Config,
	localDNSConfigurationFile string,
) (action CreateVM) {
	action.stemcellService = stemcellService
	action.virtualGuestService = virtualGuestService
	action.registryClient = registryClient
	action.registryOptions = registryOptions
	action.agentOptions = agentOptions
	action.softlayerOptions = softlayerOptions
	action.localDNSConfigFile = localDNSConfigurationFile
	return
}

func (cv CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	// Validate VM properties
	if err := cloudProps.Validate(); err != nil {
		return "", bosherr.WrapError(err, "Creating VM")
	}

	// Find stemcell uuid
	stemcellUuid, err := cv.stemcellService.Find(int(stemcellCID))
	if err != nil {
		if _, ok := err.(api.CloudError); ok {
			return "", err
		}
		return "", bosherr.WrapErrorf(err, "Finding stemcell uuid with id '%d'", stemcellCID.Int())
	}

	// Set public key
	var sshKey int
	if len(cv.softlayerOptions.PublicKey) > 0 {
		sshKey, err = cv.virtualGuestService.CreateSshKey("bosh_cpi", cv.softlayerOptions.PublicKey, cv.softlayerOptions.PublicKeyFingerPrint)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Creating Public Key with content '%s'", cv.softlayerOptions.PublicKey)
		}
		cloudProps.SshKey = sshKey
	}

	// Set userData without server name
	userData, err := cv.createUserDataForInstance(&cv.registryOptions, cloudProps.DeployedByBoshCLI)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM UserData")
	}

	// Inspect networks to get NetworkComponents
	publicNetworkComponent, privateNetworkComponent, err := cv.getNetworkComponents(networks)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting NetworkComponents from networks settings")
	}

	// Create Virtual Guest template
	virtualGuestTemplate := cv.createVirtualGuestTemplate(stemcellUuid, *cloudProps.AsInstanceProperties(), publicNetworkComponent, privateNetworkComponent)

	// Parse networks
	var instanceNetworks instance.Networks
	if publicNetworkComponent != nil {
		instanceNetworks = networks.AsInstanceServiceNetworks(publicNetworkComponent.NetworkVlan)
	} else {
		instanceNetworks = networks.AsInstanceServiceNetworks(&datatypes.Network_Vlan{})
	}

	if err = instanceNetworks.Validate(); err != nil {
		return "", bosherr.WrapError(err, "Creating VM")
	}

	if boshenv, ok := env["bosh"]; ok {
		boshenv.(map[string]interface{})["keep_root_password"] = true

		// #148050011: Set vcap password in env.bosh.password through CPI
		if env["bosh"].(map[string]interface{})["password"] == nil && cv.agentOptions.VcapPassword != "" {
			env["bosh"].(map[string]interface{})["password"] = cv.agentOptions.VcapPassword
		}
	}

	// CID for returned VM
	cid := 0
	osReloaded := false

	if !cv.softlayerOptions.DisableOsReload {
		cid, err = cv.createByOsReload(stemcellCID, virtualGuestTemplate, instanceNetworks, userData)
		if err != nil {
			return "", bosherr.WrapError(err, "OS reloading VM")
		}

		osReloaded = true
	}

	if cid == 0 {
		// Create VM
		cid, err = cv.virtualGuestService.Create(virtualGuestTemplate, cv.softlayerOptions.EnableVps, stemcellCID.Int(), []int{cloudProps.SshKey}, userData)
		if err != nil {
			if _, ok := err.(api.CloudError); ok {
				return "", err
			}
			return "", bosherr.WrapError(err, "Creating VM")
		}
	}

	instanceID := VMCID(cid).String()

	// If any of the below code fails, we must delete the created VM
	defer func() {
		if err != nil && !osReloaded {
			// #nosec G104
			cv.virtualGuestService.CleanUp(cid)
		}
	}()

	// Config VM network settings
	instanceNetworks, err = cv.virtualGuestService.ConfigureNetworks(cid, instanceNetworks)
	if err != nil {
		return "", bosherr.WrapError(err, "Configuring VM networks")
	}

	// Create VM agent settings
	agentNetworks := instanceNetworks.AsRegistryNetworks()

	// Get object details of new VM
	virtualGuest, err := cv.virtualGuestService.Find(cid)
	if err != nil {
		return "", err
	}

	if err = cv.updateHost(cid, cloudProps.DeployedByBoshCLI, networks.HasManualNetwork(), *virtualGuest.PrimaryBackendIpAddress, *virtualGuest.FullyQualifiedDomainName); err != nil {
		return "", bosherr.WrapError(err, "Post config")
	}

	agentSettings := registry.NewAgentSettings(agentID, instanceID, agentNetworks, registry.EnvSettings(env), cv.agentOptions)

	// Attach Ephemeral Disk
	if cloudProps.EphemeralDiskSize > 0 {
		err = cv.virtualGuestService.AttachEphemeralDisk(cid, cloudProps.EphemeralDiskSize)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Attaching ephemeral disk to VM with id '%d'", cid)
		}
		// Update VM agent settings
		agentSettings = agentSettings.AttachEphemeralDisk(registry.DefaultEphemeralDisk)
	}

	if err = cv.registryClient.Update(instanceID, agentSettings); err != nil {
		return "", bosherr.WrapError(err, "Updating registryClient")
	}

	return instanceID, nil
}

func (cv CreateVM) createVirtualGuestTemplate(stemcellUuid string, cloudProps VMCloudProperties,
	publicNetworkComponent *datatypes.Virtual_Guest_Network_Component, privateNetworkComponent *datatypes.Virtual_Guest_Network_Component) *datatypes.Virtual_Guest {

	virtualGuestTemplate := &datatypes.Virtual_Guest{
		// instance type
		Hostname: sl.String(cloudProps.Hostname),
		Domain:   sl.String(cloudProps.Domain),

		// datacenter or availbility zone
		Datacenter: &datatypes.Location{
			Name: sl.String(cloudProps.Datacenter),
		},

		// stemcell or image
		BlockDeviceTemplateGroup: &datatypes.Virtual_Guest_Block_Device_Template_Group{
			GlobalIdentifier: sl.String(stemcellUuid),
		},

		// billing options
		HourlyBillingFlag: sl.Bool(cloudProps.HourlyBillingFlag),
		LocalDiskFlag:     sl.Bool(cloudProps.LocalDiskFlag),

		// network components
		NetworkComponents: []datatypes.Virtual_Guest_Network_Component{
			{MaxSpeed: sl.Int(cloudProps.MaxNetworkSpeed)},
		},
		PrivateNetworkOnlyFlag:         sl.Bool(publicNetworkComponent == nil),
		PrimaryNetworkComponent:        publicNetworkComponent,
		PrimaryBackendNetworkComponent: privateNetworkComponent,
	}

	if cloudProps.FlavorKeyName != "" {
		virtualGuestTemplate.SupplementalCreateObjectOptions = &datatypes.Virtual_Guest_SupplementalCreateObjectOptions{
			FlavorKeyName: sl.String(cloudProps.FlavorKeyName),
		}
	} else {
		virtualGuestTemplate.StartCpus = sl.Int(cloudProps.Cpu)
		virtualGuestTemplate.MaxMemory = sl.Int(cloudProps.Memory)
	}

	if cloudProps.DedicatedAccountHostOnlyFlag {
		virtualGuestTemplate.DedicatedAccountHostOnlyFlag = sl.Bool(cloudProps.DedicatedAccountHostOnlyFlag)
	} else if cloudProps.DedicatedHostId != 0 {
		virtualGuestTemplate.DedicatedHost = &datatypes.Virtual_DedicatedHost{
			Id: sl.Int(cloudProps.DedicatedHostId),
		}
	}

	if cloudProps.SshKey != 0 {
		virtualGuestTemplate.SshKeys = []datatypes.Security_Ssh_Key{{Id: sl.Int(cloudProps.SshKey)}}
	}

	return virtualGuestTemplate
}

func (cv CreateVM) getNetworkComponents(networks Networks) (*datatypes.Virtual_Guest_Network_Component, *datatypes.Virtual_Guest_Network_Component, error) {
	var publicNetworkComponent, privateNetworkComponent *datatypes.Virtual_Guest_Network_Component

	for name, nw := range networks {
		if nw.Type == "manual" {
			continue
		}
		if len(nw.CloudProperties.SubnetIds) > 0 {
			for _, subnetId := range nw.CloudProperties.SubnetIds {
				networkComponent, err := cv.createNetworkComponentsBySubnetId(subnetId)
				if err != nil {
					return &datatypes.Virtual_Guest_Network_Component{},
						&datatypes.Virtual_Guest_Network_Component{},
						bosherr.WrapErrorf(err, "Network: %s, subnet id: %d", name, subnetId)
				}

				switch *networkComponent.NetworkVlan.NetworkSpace {
				case "PRIVATE":
					if privateNetworkComponent == nil {
						privateNetworkComponent = networkComponent
					} else if privateNetworkComponent.NetworkVlan.PrimarySubnetId != networkComponent.NetworkVlan.PrimarySubnetId {
						return &datatypes.Virtual_Guest_Network_Component{},
							&datatypes.Virtual_Guest_Network_Component{},
							bosherr.Error("Only one private VLAN is supported")
					}
				case "PUBLIC":
					if publicNetworkComponent == nil {
						publicNetworkComponent = networkComponent
					} else if publicNetworkComponent.NetworkVlan.PrimarySubnetId != publicNetworkComponent.NetworkVlan.PrimarySubnetId {
						return &datatypes.Virtual_Guest_Network_Component{},
							&datatypes.Virtual_Guest_Network_Component{},
							bosherr.Error("Only one public VLAN is supported")
					}
				default:
					return &datatypes.Virtual_Guest_Network_Component{},
						&datatypes.Virtual_Guest_Network_Component{},
						bosherr.Errorf("networkVlan %d: unknown network type '%s'", subnetId, *networkComponent.NetworkVlan.NetworkSpace)
				}
			}
		} else if len(nw.CloudProperties.VlanIds) > 0 {
			for _, vlanId := range nw.CloudProperties.VlanIds {
				networkComponent, err := cv.createNetworkComponentsByVlanId(vlanId)
				if err != nil {
					return &datatypes.Virtual_Guest_Network_Component{},
						&datatypes.Virtual_Guest_Network_Component{},
						bosherr.WrapErrorf(err, "Network: %s, vlan id: %d", name, vlanId)
				}

				switch *networkComponent.NetworkVlan.NetworkSpace {
				case "PRIVATE":
					if privateNetworkComponent == nil {
						privateNetworkComponent = networkComponent
					} else if privateNetworkComponent.NetworkVlan.Id != networkComponent.NetworkVlan.Id {
						return &datatypes.Virtual_Guest_Network_Component{},
							&datatypes.Virtual_Guest_Network_Component{},
							bosherr.Error("Only one private VLAN is supported")
					}
				case "PUBLIC":
					if publicNetworkComponent == nil {
						publicNetworkComponent = networkComponent
					} else if publicNetworkComponent.NetworkVlan.Id != publicNetworkComponent.NetworkVlan.Id {
						return &datatypes.Virtual_Guest_Network_Component{},
							&datatypes.Virtual_Guest_Network_Component{},
							bosherr.Error("Only one public VLAN is supported")
					}
				default:
					return &datatypes.Virtual_Guest_Network_Component{},
						&datatypes.Virtual_Guest_Network_Component{},
						bosherr.Errorf("networkVlan %d: unknown network type '%s'", vlanId, *networkComponent.NetworkVlan.NetworkSpace)
				}
			}
		}
	}

	if privateNetworkComponent == nil {
		return publicNetworkComponent, privateNetworkComponent, bosherr.Error("A private network is required. Please check vlan_ids")
	}

	return publicNetworkComponent, privateNetworkComponent, nil
}

func (cv CreateVM) createNetworkComponentsBySubnetId(subnetId int) (*datatypes.Virtual_Guest_Network_Component, error) {
	subnet, err := cv.virtualGuestService.GetSubnet(subnetId, boslc.NETWORK_DEFAULT_SUBNET_MASK)
	if err != nil {
		return &datatypes.Virtual_Guest_Network_Component{}, bosherr.WrapErrorf(err, "Getting subnet info with id '%d'", subnetId)
	}
	if subnetId != *subnet.Id {
		return &datatypes.Virtual_Guest_Network_Component{}, bosherr.WrapErrorf(err,
			"SubnetId '%d' is not suitable with subnet '%d'", subnetId, subnet.Id)
	}
	return &datatypes.Virtual_Guest_Network_Component{
		NetworkVlan: &datatypes.Network_Vlan{
			PrimarySubnetId: subnet.Id,
			NetworkSpace:    subnet.AddressSpace,
		},
	}, nil
}

func (cv CreateVM) createNetworkComponentsByVlanId(vlanId int) (*datatypes.Virtual_Guest_Network_Component, error) {
	vlan, err := cv.virtualGuestService.GetVlan(vlanId, boslc.NETWORK_DEFAULT_VLAN_MASK)
	if err != nil {
		return &datatypes.Virtual_Guest_Network_Component{}, bosherr.WrapErrorf(err, "Getting vlan info with id '%d'", vlanId)
	}
	if vlanId != *(vlan.Id) {
		return &datatypes.Virtual_Guest_Network_Component{}, bosherr.WrapErrorf(err,
			"VlanId '%d' is not suitable with vlan '%d'", vlanId, *(vlan.Id))
	}
	return &datatypes.Virtual_Guest_Network_Component{
		NetworkVlan: &datatypes.Network_Vlan{
			Id:           vlan.Id,
			NetworkSpace: vlan.NetworkSpace,
		},
	}, nil
}

func (cv CreateVM) createByOsReload(stemcellCID StemcellCID, template *datatypes.Virtual_Guest, instanceNetworks instance.Networks, userData *registry.SoftlayerUserData) (int, error) {
	cid := 0
	for _, network := range instanceNetworks {
		switch network.Type {
		case "dynamic":
			if len(network.IP) > 0 && cid == 0 {
				var (
					vm            *datatypes.Virtual_Guest
					sshKey        int
					desiredCpu    int
					desiredMemory int
					err           error
				)

				if IsPrivateSubnet(net.ParseIP(network.IP)) {
					vm, err = cv.virtualGuestService.FindByPrimaryBackendIp(network.IP)
				} else {
					vm, err = cv.virtualGuestService.FindByPrimaryIp(network.IP)
				}
				if err != nil {
					return cid, err
				}

				if err := cv.registryClient.Delete(strconv.Itoa(*vm.Id)); err != nil {
					return cid, bosherr.WrapErrorf(err, "Cleaning registry record '%d' before os_reload", *vm.Id)
				}

				if (template.SupplementalCreateObjectOptions == nil || *template.SupplementalCreateObjectOptions.FlavorKeyName == "") &&
					vm.SupplementalCreateObjectOptions == nil {
					if *vm.MaxCpu != *template.StartCpus {
						desiredCpu = sl.Get(*template.StartCpus, 0).(int)

					}
					if *vm.MaxMemory != *template.MaxMemory {
						desiredMemory = sl.Get(*template.MaxMemory, 0).(int)
					}

					if desiredCpu != 0 || desiredMemory != 0 {
						err = cv.virtualGuestService.UpgradeInstance(*vm.Id, desiredCpu, desiredMemory, 0, *vm.DedicatedAccountHostOnlyFlag, vm.DedicatedHost != nil)
						if err != nil {
							return cid, bosherr.WrapError(err, "Upgrading VM")
						}
					}
				}

				if len(template.SshKeys) != 0 {
					sshKey = *template.SshKeys[0].Id
				}

				userData.Server = registry.SoftlayerUserDataServerName{
					Name: strconv.Itoa(*vm.Id),
				}

				err = cv.virtualGuestService.ReloadOS(*vm.Id, stemcellCID.Int(), []int{sshKey}, *template.Hostname, *template.Domain, userData)
				if err != nil {
					return cid, err
				}

				cid = *vm.Id
			}
		case "vip":
			return cid, bosherr.Error("SoftLayer Not Support VIP Networking")
		default:
			continue
		}
	}

	return cid, nil
}

func (cv CreateVM) createUserDataForInstance(registryOptions *registry.ClientOptions, deployedByBoshCLI bool) (*registry.SoftlayerUserData, error) {
	var directorIP string
	var err error
	if deployedByBoshCLI {
		directorIP = "127.0.0.1"
	} else {
		directorIP, err = cv.getDirectorIPAddressByHost(registryOptions.Address)
		if err != nil {
			return &registry.SoftlayerUserData{}, bosherr.WrapError(err, "Failed to get bosh director IP address in local")
		}
	}
	registryOptions.Address = directorIP
	userDataContents := registry.SoftlayerUserData{
		Registry: registry.SoftlayerUserDataRegistryEndpoint{
			Endpoint: fmt.Sprintf("http://%s:%s@%s:%d",
				registryOptions.HTTPOptions.User,
				registryOptions.HTTPOptions.Password,
				registryOptions.Address,
				registryOptions.HTTPOptions.Port),
		},
	}

	return &userDataContents, nil
}

func (cv CreateVM) updateHostsFile(newIpAddress string, targetHostname string) (err error) {
	err = os.Setenv("HOSTS_PATH", cv.localDNSConfigFile)
	if err != nil {
		return bosherr.WrapErrorf(err, "Set '%s' to env variable 'HOSTS_PATH'", cv.localDNSConfigFile)
	}
	hosts, err := goodhosts.NewHosts()
	if err != nil {
		return bosherr.WrapErrorf(err, "Load hosts file")
	}
	err = hosts.RemoveByHostname(targetHostname)
	if err != nil {
		return bosherr.WrapErrorf(err, "Remove '%s' in hosts", targetHostname)
	}
	err = hosts.Add(newIpAddress, targetHostname)
	if err != nil {
		return bosherr.WrapErrorf(err, "Add '%s %s' in hosts", newIpAddress, targetHostname)
	}

	if err := hosts.Flush(); err != nil {
		return bosherr.WrapErrorf(err, "Flush hosts file")
	}

	return nil
}

func (cv CreateVM) getDirectorIPAddressByHost(host string) (string, error) {
	// check host is ip address or hostname
	address := net.ParseIP(host)
	if address == nil || address.String() == "127.0.0.1" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Failed to get network interfaces")
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}

		return "", bosherr.Error(fmt.Sprintf("Failed to get IP address by '%s'", host))
	}

	return address.String(), nil
}

func (cv CreateVM) postConfig(virtualGuestId int, agentOptions *registry.AgentOptions) error {
	mbus, err := cv.parseMbusURL(agentOptions.Mbus)
	if err != nil {
		return bosherr.WrapError(err, "Cannot construct mbus url.")
	}
	agentOptions.Mbus = mbus

	switch agentOptions.Blobstore.Provider {
	case "dav":
		davConf := instance.DavConfig(agentOptions.Blobstore.Options)
		err = cv.updateDavConfig(&davConf)
		if err != nil {
			return bosherr.WrapError(err, "Cannot update dav config.")
		}
	}
	return nil
}

func (cv CreateVM) parseMbusURL(mbusURL string) (string, error) {
	parsedURL, err := url.Parse(mbusURL)
	if err != nil {
		return "", bosherr.WrapError(err, "Parsing Mbus URL")
	}

	var username, password, port string
	host, port, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		return "", bosherr.WrapError(err, "Spliting host and port")
	}

	ipAddress, err := cv.getDirectorIPAddressByHost(host)

	userInfo := parsedURL.User
	if userInfo != nil {
		username = userInfo.Username()
		password, _ = userInfo.Password()
		return fmt.Sprintf("%s://%s:%s@%s:%s", parsedURL.Scheme, username, password, ipAddress, port), nil
	}

	return fmt.Sprintf("%s://%s:%s", parsedURL.Scheme, ipAddress, port), nil
}

func (cv CreateVM) updateDavConfig(config *instance.DavConfig) (err error) {
	url := (*config)["endpoint"].(string)
	mbus, err := cv.parseMbusURL(url)
	if err != nil {
		return bosherr.WrapError(err, "Parsing Mbus URL")
	}

	(*config)["endpoint"] = mbus

	return nil
}

func (cv *CreateVM) updateHost(cid int, deployedByBoshCLI bool, hasManualNetwork bool, ipAddress string, hostname string) error {
	if !deployedByBoshCLI {
		// Config mbus and blobstore host if needed
		if err := cv.postConfig(cid, &cv.agentOptions); err != nil {
			return bosherr.WrapError(err, "Post config")
		}

		return nil
	}

	// Only update `/etc/hosts` when director has manual network
	if hasManualNetwork {
		err := cv.updateHostsFile(ipAddress, hostname)
		if err != nil {
			return bosherr.WrapError(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	}

	return nil
}

type ipRange struct {
	start net.IP
	end   net.IP
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

func IsPrivateSubnet(ipAddress net.IP) bool {
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		for _, r := range privateRanges {
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

func inRange(r ipRange, ipAddress net.IP) bool {
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) <= 0 {
		return true
	}
	return false
}
