package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"
	"bosh-softlayer-cpi/util"

	boslc "bosh-softlayer-cpi/softlayer/client"
	boslconfig "bosh-softlayer-cpi/softlayer/config"

	"bosh-softlayer-cpi/softlayer/stemcell_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/lextoumbourou/goodhosts"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
	"net/url"
	"os"
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
	globalIdentifier, found, err := cv.stemcellService.Find(stemcellCID.Int())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating VM")
	}
	if !found {
		return "", api.NewStemcellkNotFoundError(stemcellCID.String(), false)
	}

	var sshKey int
	if len(cv.softlayerOptions.PublicKey) > 0 {
		sshKey, err = cv.virtualGuestService.CreateSshKey("bosh_cpi", cv.softlayerOptions.PublicKey, cv.softlayerOptions.PublicKeyFingerPrint)
		if err != nil {
			return "", bosherr.WrapErrorf(err, "Creating Public Key with content '%s'", cv.softlayerOptions.PublicKey)
		}
		cloudProps.SshKey = sshKey
	}

	cv.updateCloudProperties(&cloudProps)

	// create VM user data
	userDataTypeContents, err := cv.createUserDataForInstance(agentID, cv.registryOptions)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM UserData")
	}

	publicVlanId, privateVlanId, err := cv.getVlanIds(networks)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting vlan ids from networks settings")
	}

	virtualGuestTemplate := cv.createVirtualGuestTemplate(globalIdentifier, cloudProps, userDataTypeContents, publicVlanId, privateVlanId)

	// Parse networks
	instanceNetworks := networks.AsInstanceServiceNetworks()
	if err = instanceNetworks.Validate(); err != nil {
		return "", bosherr.WrapError(err, "Creating VM")
	}

	// Validate VM tags and labels
	if err = cloudProps.Validate(); err != nil {
		return "", bosherr.WrapError(err, "Creating VM")
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
		cid, err = cv.virtualGuestService.Create(*virtualGuestTemplate, instanceNetworks, cv.registryOptions.EndpointWithCredentials())
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

	// Keep root password
	env["bosh"].(map[string]interface{})["keep_root_password"] = true

	// Get object details of new VM
	virtualGuest, found, err := cv.virtualGuestService.Find(cid)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating VM")
	}
	if !found {
		return "", api.NewVMNotFoundError(string(cid))
	}

	if cloudProps.DeployedByBoshCLI {
		err := cv.updateHosts("/etc/hosts", *virtualGuest.PrimaryBackendIpAddress, *virtualGuest.FullyQualifiedDomainName)
		if err != nil {
			return "", bosherr.WrapError(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		// Post config
		if !networks.HasManualNetwork() {
			if err = cv.postConfig(cid, &cv.agentOptions); err != nil {
				return "", bosherr.WrapError(err, "Post config")
			}
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

func (cv CreateVM) updateCloudProperties(cloudProps *VMCloudProperties) {
	if cloudProps.DeployedByBoshCLI {
		cloudProps.VmNamePrefix = cv.updateHostNameInCloudProps(cloudProps, "")
	} else {
		cloudProps.VmNamePrefix = cv.updateHostNameInCloudProps(cloudProps, cv.timeStampForTime(time.Now().UTC()))
	}

	if cloudProps.StartCpus == 0 {
		cloudProps.StartCpus = 4
	}

	if cloudProps.MaxMemory == 0 {
		cloudProps.MaxMemory = 8192
	}

	if len(cloudProps.Domain) == 0 {
		cloudProps.Domain = "softlayer.com"
	}

	if cloudProps.MaxNetworkSpeed == 0 {
		cloudProps.MaxNetworkSpeed = 1000
	}
}

func (cv CreateVM) updateHostNameInCloudProps(cloudProps *VMCloudProperties, timeStampPostfix string) string {
	if len(timeStampPostfix) == 0 {
		return cloudProps.VmNamePrefix
	} else {
		return cloudProps.VmNamePrefix + "-" + timeStampPostfix
	}
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

func (cv CreateVM) createVirtualGuestTemplate(stemcellUuid string, cloudProps VMCloudProperties, userData string, publicVlanId int, privateVlanId int) *datatypes.Virtual_Guest {
	var publicNetworkComponent, privateNetworkComponent *datatypes.Virtual_Guest_Network_Component

	if publicVlanId != 0 {
		publicNetworkComponent = &datatypes.Virtual_Guest_Network_Component{
			NetworkVlan: &datatypes.Network_Vlan{
				Id: sl.Int(publicVlanId),
			},
		}
	}

	privateNetworkComponent = &datatypes.Virtual_Guest_Network_Component{
		NetworkVlan: &datatypes.Network_Vlan{
			Id: sl.Int(privateVlanId),
		},
	}

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

func (cv CreateVM) createByOsReload(stemcellCID StemcellCID, cloudProps VMCloudProperties, instanceNetworks instance.Networks) (int, error) {
	cid := 0
	for _, network := range instanceNetworks {
		switch network.Type {
		case "dynamic":
			if len(network.IP) > 0 {
				if util.IsPrivateSubnet(net.ParseIP(network.IP)) {
					vm, found, err := cv.virtualGuestService.FindByPrimaryBackendIp(network.IP)
					if err != nil {
						return cid, bosherr.WrapErrorf(err, "Finding VM with IP Address '%s'", network.IP)
					}
					if !found {
						return cid, api.NewVMCreationFailedError(fmt.Sprintf("Finding VM with IP Address '%s'", network.IP), true)
					}

					_, err = cv.virtualGuestService.ReloadOS(*vm.Id, stemcellCID.Int(), []int{cloudProps.SshKey})
					if err != nil {
						if apiErr, ok := err.(sl.Error); ok {
							return cid, api.NewVMCreationFailedError(fmt.Sprintf("Failed to do OS Reload with IP Address '%s' with error %s", network.IP, apiErr), false)
						} else {
							return cid, api.NewVMCreationFailedError(fmt.Sprintf("Failed to do OS Reload with IP Address '%s' with error %s", network.IP, apiErr), true)
						}
					}

					cid = *vm.Id

					succeed, err := cv.virtualGuestService.Edit(*vm.Id, datatypes.Virtual_Guest{
						Hostname: sl.String(cloudProps.VmNamePrefix),
						Domain:   sl.String(cloudProps.Domain),
					})
					if err != nil {
						return cid, api.NewVMCreationFailedError(fmt.Sprintf("Editing VM hostname after OS Reload with IP Address '%s' with error %s", network.IP, err), true)
					}

					if !succeed {

						return cid, api.NewVMCreationFailedError(fmt.Sprintf("Failed to edit VM hostname after OS Reload with IP Address '%s'", network.IP), true)
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

func (cv CreateVM) getVlanIds(networks Networks) (int, int, error) {
	var publicVlanID, privateVlanID int

	for name, nw := range networks {
		if nw.Type == "manual" {
			continue
		}
		for _, vlanId := range nw.CloudProperties.VlanIds {
			networkSpace, err := cv.getNetworkSpace(vlanId)
			if err != nil {
				return 0, 0, bosherr.WrapErrorf(err, "Network: %q, vlan id: %d", name, vlanId)
			}

			switch networkSpace {
			case "PRIVATE":
				if privateVlanID == 0 {
					privateVlanID = vlanId
				} else if privateVlanID != vlanId {
					return 0, 0, bosherr.Error("Only one private VLAN is supported")
				}
			case "PUBLIC":
				if publicVlanID == 0 {
					publicVlanID = vlanId
				} else if publicVlanID != vlanId {
					return 0, 0, bosherr.Error("Only one public VLAN is supported")
				}
			default:
				return 0, 0, bosherr.Errorf("Vlan id %d: unknown network type '%s'", vlanId, networkSpace)
			}
		}
	}

	if privateVlanID == 0 {
		return 0, 0, bosherr.Error("A private vlan is required")
	}

	return publicVlanID, privateVlanID, nil
}

func (cv CreateVM) getNetworkSpace(vlanID int) (string, error) {
	networkVlan, err := cv.virtualGuestService.GetVlan(vlanID, boslc.NETWORK_DEFAULT_VLAN)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Getting vlan info with id '%d'", vlanID)
	}
	return *networkVlan.NetworkSpace, nil
}

func (cv CreateVM) timeStampForTime(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + fmt.Sprintf("%03d", int(now.UnixNano()/1e6-now.Unix()*1e3))
}

func (cv CreateVM) updateHosts(path string, newIpAddress string, targetHostname string) (err error) {
	err = os.Setenv("HOSTS_PATH", path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Set '%s' to env variable 'HOSTS_PATH'", path)
	}
	hosts, err := goodhosts.NewHosts()
	if err != nil {
		return bosherr.WrapErrorf(err, "Load hosts file")
	}
	matchedAddresses, err := net.LookupHost(targetHostname)
	for _, address := range matchedAddresses {
		if hosts.Has(address, targetHostname) {
			err := hosts.Remove(address, targetHostname)
			if err != nil {
				return bosherr.WrapErrorf(err, "Remove '%s %s' in hosts", address, targetHostname)
			}
		}
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

func (cv CreateVM) parseMbusURL(mbusURL string, primaryBackendIpAddress string) (string, error) {
	parsedURL, err := url.Parse(mbusURL)
	if err != nil {
		return "", bosherr.WrapError(err, "Parsing Mbus URL")
	}

	var username, password, port string
	_, port, _ = net.SplitHostPort(parsedURL.Host)
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
