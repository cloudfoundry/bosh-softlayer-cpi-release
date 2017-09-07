package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"

	boslc "bosh-softlayer-cpi/softlayer/client"
	boslconfig "bosh-softlayer-cpi/softlayer/config"

	"bosh-softlayer-cpi/softlayer/stemcell_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	"bytes"
	"github.com/bluebosh/goodhosts"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
	"net/url"
	"os"
	"strconv"
)

type CreateVM struct {
	stemcellService     stemcell.Service
	virtualGuestService instance.Service
	registryClient      registry.Client
	registryOptions     registry.ClientOptions
	agentOptions        registry.AgentOptions
	softlayerOptions    boslconfig.Config
}

func NewCreateVM(
	stemcellService stemcell.Service,
	virtualGuestService instance.Service,
	registryClient registry.Client,
	registryOptions registry.ClientOptions,
	agentOptions registry.AgentOptions,
	softlayerOptions boslconfig.Config,
) (action CreateVM) {
	action.stemcellService = stemcellService
	action.virtualGuestService = virtualGuestService
	action.registryClient = registryClient
	action.registryOptions = registryOptions
	action.agentOptions = agentOptions
	action.softlayerOptions = softlayerOptions
	return
}

func (cv CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	// Validate VM properties
	if err := cloudProps.Validate(); err != nil {
		return "", bosherr.WrapError(err, "Creating VM")
	}

	// Find stemcell uuid
	globalIdentifier, err := cv.stemcellService.Find(int(stemcellCID))
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

	// Create VM user data
	userDataContents, err := cv.createUserDataForInstance(agentID, cv.registryOptions)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM UserData")
	}

	// Inspect networks to get NetworkComponents
	//publicVlanId, privateVlanId, err := cv.getVlanIds(networks)
	publicNetworkComponent, privateNetworkComponent, err := cv.getNetworkComponents(networks)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting NetworkComponents from networks settings")
	}

	// Create Virtual Guest template
	virtualGuestTemplate := cv.createVirtualGuestTemplate(globalIdentifier, *cloudProps.AsInstanceProperties(), userDataContents, publicNetworkComponent, privateNetworkComponent)

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

	// Extract any tags from env.bosh.groups
	if boshenv, ok := env["bosh"]; ok {
		if boshgroups, ok := boshenv.(map[string]interface{})["groups"]; ok {
			for _, tag := range boshgroups.([]interface{}) {
				// Ignore error as labels will be validated later
				cloudProps.Tags = append(cloudProps.Tags, tag.(string))
			}
		}

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
		cid, err = cv.createByOsReload(stemcellCID, cloudProps, instanceNetworks)
		if err != nil {
			return "", bosherr.WrapError(err, "OS reloading VM")
		}

		osReloaded = true
	}

	if cid == 0 {
		// Create VM
		cid, err = cv.virtualGuestService.Create(virtualGuestTemplate, cv.softlayerOptions.EnableVps, stemcellCID.Int(), []int{cloudProps.SshKey})
		if err != nil {
			if _, ok := err.(api.CloudError); ok {
				return "", err
			}
			return "", bosherr.WrapError(err, "Creating VM")
		}
	}

	// If any of the below code fails, we must delete the created cid
	defer func() {
		if err != nil && !osReloaded {
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

	if cloudProps.DeployedByBoshCLI {
		err := cv.updateHosts("/etc/hosts", *virtualGuest.PrimaryBackendIpAddress, *virtualGuest.FullyQualifiedDomainName)
		if err != nil {
			return "", bosherr.WrapError(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		// Config mbus and blobstore options if needed
		if err = cv.postConfig(cid, &cv.agentOptions); err != nil {
			return "", bosherr.WrapError(err, "Post config")
		}
	}

	agentSettings := registry.NewAgentSettings(agentID, VMCID(cid).String(), agentNetworks, registry.EnvSettings(env), cv.agentOptions)

	// Attach Ephemeral Disk
	if cloudProps.EphemeralDiskSize > 0 {
		err = cv.virtualGuestService.AttachEphemeralDisk(cid, cloudProps.EphemeralDiskSize)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Attaching ephemeral disk to VM with id '%d'", cid)
		}
		//Update VM agent settings
		agentSettings = agentSettings.AttachEphemeralDisk(registry.DefaultEphemeralDisk)
	}

	if err = cv.registryClient.Update(VMCID(cid).String(), agentSettings); err != nil {
		return "", bosherr.WrapError(err, "Updating registryClient")
	}

	return VMCID(cid).String(), nil
}

func (cv CreateVM) createVirtualGuestTemplate(stemcellUuid string, cloudProps VMCloudProperties, userData string,
	publicNetworkComponent *datatypes.Virtual_Guest_Network_Component, privateNetworkComponent *datatypes.Virtual_Guest_Network_Component) *datatypes.Virtual_Guest {

	virtualGuestTemplate := &datatypes.Virtual_Guest{
		// instance type
		Hostname:  sl.String(cloudProps.VmNamePrefix),
		Domain:    sl.String(cloudProps.Domain),
		StartCpus: sl.Int(cloudProps.StartCpus),
		MaxMemory: sl.Int(cloudProps.MaxMemory),

		// datacenter or availbility zone
		Datacenter: &datatypes.Location{
			Name: sl.String(cloudProps.Datacenter),
		},

		// stemcell or image
		BlockDeviceTemplateGroup: &datatypes.Virtual_Guest_Block_Device_Template_Group{
			GlobalIdentifier: sl.String(stemcellUuid),
		},

		// billing options
		HourlyBillingFlag:            sl.Bool(cloudProps.HourlyBillingFlag),
		LocalDiskFlag:                sl.Bool(cloudProps.LocalDiskFlag),
		DedicatedAccountHostOnlyFlag: sl.Bool(cloudProps.DedicatedAccountHostOnlyFlag),

		// network components
		NetworkComponents: []datatypes.Virtual_Guest_Network_Component{
			{MaxSpeed: sl.Int(cloudProps.MaxNetworkSpeed)},
		},
		PrivateNetworkOnlyFlag:         sl.Bool(publicNetworkComponent == nil),
		PrimaryNetworkComponent:        publicNetworkComponent,
		PrimaryBackendNetworkComponent: privateNetworkComponent,

		// metadata or user data
		UserData: []datatypes.Virtual_Guest_Attribute{
			{
				Value: sl.String(userData),
			},
		},
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

func (cv CreateVM) createByOsReload(stemcellCID StemcellCID, cloudProps VMCloudProperties, instanceNetworks instance.Networks) (int, error) {
	cid := 0
	for _, network := range instanceNetworks {
		switch network.Type {
		case "dynamic":
			if len(network.IP) > 0 && cid == 0 {
				var (
					vm  *datatypes.Virtual_Guest
					err error
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

				err = cv.virtualGuestService.ReloadOS(*vm.Id, stemcellCID.Int(), []int{cloudProps.SshKey}, cloudProps.VmNamePrefix, cloudProps.Domain)
				if err != nil {
					return cid, err
				}

				cid = *vm.Id

				if *vm.MaxCpu != cloudProps.StartCpus ||
					*vm.MaxMemory != cloudProps.MaxMemory ||
					*vm.DedicatedAccountHostOnlyFlag != cloudProps.DedicatedAccountHostOnlyFlag {
					err = cv.virtualGuestService.UpgradeInstance(cid, cloudProps.StartCpus, cloudProps.MaxMemory, 0, cloudProps.DedicatedAccountHostOnlyFlag)
					if err != nil {
						return cid, bosherr.WrapError(err, "Upgrading VM")
					}
				}
			}
		case "vip":
			return cid, bosherr.Error("SoftLayer Not Support VIP Networking")
		default:
			continue
		}
	}

	return cid, nil
}

func (cv CreateVM) createUserDataForInstance(agentID string, registryOptions registry.ClientOptions) (string, error) {
	serverName := fmt.Sprintf("vm-%s", agentID)
	userDataContents := registry.SoftlayerUserData{
		Registry: registry.SoftlayerUserDataRegistryEndpoint{
			Endpoint: fmt.Sprintf("http://%s:%s@%s:%d",
				registryOptions.HTTPOptions.User,
				registryOptions.HTTPOptions.Password,
				registryOptions.Host,
				registryOptions.HTTPOptions.Port),
		},
		Server: registry.SoftlayerUserDataServerName{
			Name: serverName,
		},
	}
	contentsBytes, err := json.Marshal(userDataContents)
	if err != nil {
		return "", bosherr.WrapError(err, "Preparing user data contents")
	}

	return base64.RawURLEncoding.EncodeToString(contentsBytes), nil
}

//func (cv CreateVM) getVlanIds(networks Networks) (int, int, error) {
//	var publicVlanID, privateVlanID int
//
//	for name, nw := range networks {
//		if nw.Type == "manual" {
//			continue
//		}
//		for _, vlanId := range nw.CloudProperties.VlanIds {
//			networkSpace, err := cv.getNetworkSpace(vlanId)
//			if err != nil {
//				return 0, 0, bosherr.WrapErrorf(err, "Network: %q, vlan id: %d", name, vlanId)
//			}
//
//			switch networkSpace {
//			case "PRIVATE":
//				if privateVlanID == 0 {
//					privateVlanID = vlanId
//				} else if privateVlanID != vlanId {
//					return 0, 0, bosherr.Error("Only one private VLAN is supported")
//				}
//			case "PUBLIC":
//				if publicVlanID == 0 {
//					publicVlanID = vlanId
//				} else if publicVlanID != vlanId {
//					return 0, 0, bosherr.Error("Only one public VLAN is supported")
//				}
//			default:
//				return 0, 0, bosherr.Errorf("Vlan id %d: unknown network type '%s'", vlanId, networkSpace)
//			}
//		}
//	}
//
//	if privateVlanID == 0 {
//		return 0, 0, bosherr.Error("A private vlan is required")
//	}
//
//	return publicVlanID, privateVlanID, nil
//}
//
//func (cv CreateVM) getNetworkSpace(vlanID int) (string, error) {
//	networkVlan, err := cv.virtualGuestService.GetVlan(vlanID, boslc.NETWORK_DEFAULT_VLAN)
//	if err != nil {
//		return "", bosherr.WrapErrorf(err, "Getting vlan info with id '%d'", vlanID)
//	}
//	return *networkVlan.NetworkSpace, nil
//}

func (cv CreateVM) updateHosts(path string, newIpAddress string, targetHostname string) (err error) {
	err = os.Setenv("HOSTS_PATH", path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Set '%s' to env variable 'HOSTS_PATH'", path)
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

func (cv CreateVM) getLocalIPAddress() (string, error) {
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

	return "", bosherr.Error(fmt.Sprintf("Failed to get IP address"))
}

func (cv CreateVM) postConfig(virtualGuestId int, agentOptions *registry.AgentOptions) error {
	boshIP, err := cv.getLocalIPAddress()
	if err != nil {
		return bosherr.WrapError(err, "Failed to get IP address in local")
	}

	mbus, err := cv.parseMbusURL(agentOptions.Mbus, boshIP)
	if err != nil {
		return bosherr.WrapError(err, "Cannot construct mbus url.")
	}
	agentOptions.Mbus = mbus

	switch agentOptions.Blobstore.Provider {
	case "dav":
		davConf := instance.DavConfig(agentOptions.Blobstore.Options)
		cv.updateDavConfig(&davConf, boshIP)
	}
	return nil
}

func (cv CreateVM) parseMbusURL(mbusURL string, primaryBackendIpAddress string) (string, error) {
	parsedURL, err := url.Parse(mbusURL)
	if err != nil {
		return "", bosherr.WrapError(err, "Parsing Mbus URL")
	}

	var username, password, port string
	host, port, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		return "", bosherr.WrapError(err, "Spliting host and port")
	}

	// when director using dynamic network, the mbus host should be "0.0.0.0"
	if host != "0.0.0.0" {
		return mbusURL, nil
	}

	userInfo := parsedURL.User
	if userInfo != nil {
		username = userInfo.Username()
		password, _ = userInfo.Password()
		return fmt.Sprintf("%s://%s:%s@%s:%s", parsedURL.Scheme, username, password, primaryBackendIpAddress, port), nil
	}

	return fmt.Sprintf("%s://%s:%s", parsedURL.Scheme, primaryBackendIpAddress, port), nil
}

func (cv CreateVM) updateDavConfig(config *instance.DavConfig, directorIP string) (err error) {
	url := (*config)["endpoint"].(string)
	mbus, err := cv.parseMbusURL(url, directorIP)
	if err != nil {
		return bosherr.WrapError(err, "Parsing Mbus URL")
	}

	(*config)["endpoint"] = mbus

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
