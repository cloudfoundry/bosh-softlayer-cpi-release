package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/registry"
	boslc "bosh-softlayer-cpi/softlayer/client"

	"bosh-softlayer-cpi/softlayer/stemcell_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

type CreateVM struct {
	stemcellService     stemcell.Service
	virtualGuestService instance.Service
	registryClient      registry.Client
	registryOptions     registry.ClientOptions
	agentOptions        registry.AgentOptions
}

func NewCreateVM(
	stemcellService stemcell.Service,
	virtualGuestService instance.Service,
	registryClient registry.Client,
	registryOptions registry.ClientOptions,
	agentOptions registry.AgentOptions,
) (action CreateVM) {
	action.stemcellService = stemcellService
	action.virtualGuestService = virtualGuestService
	action.registryClient = registryClient
	action.registryOptions = registryOptions
	action.agentOptions = agentOptions
	return
}

func (cv CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks Networks, diskIDs []DiskCID, env Environment) (string, error) {
	globalIdentifier, found, err := cv.stemcellService.Find(stemcellCID.Int())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating VM")
	}
	if !found {
		return "", api.NewStemcellkNotFoundError(string(stemcellCID), false)
	}

	cv.updateCloudProperties(&cloudProps)

	userDataTypeContents, err := cv.createUserDataForInstance(agentID, cv.registryOptions)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM userData")
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

	// Parse VM properties
	vmProps := &instance.Properties{
		VirtualGuestTemplate: *virtualGuestTemplate,
		SecondDisk:           cloudProps.EphemeralDiskSize,
		DeployedByBoshCLI:    cloudProps.DeployedByBoshCLI,
	}

	// Create VM
	cid, err := cv.virtualGuestService.Create(vmProps, instanceNetworks, cv.registryOptions.EndpointWithCredentials())
	if err != nil {
		if _, ok := err.(api.CloudError); ok {
			return "", err
		}
		return "", bosherr.WrapError(err, "Creating VM")
	}

	// If any of the below code fails, we must delete the created cid
	defer func() {
		if err != nil {
			cv.virtualGuestService.CleanUp(cid)
		}
	}()

	// Config VM network settings

	instanceNetworks, err = cv.virtualGuestService.ConfigureNetworks(cid, instanceNetworks)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Configuring VM networks")
	}

	// Create VM settings
	agentNetworks := instanceNetworks.AsRegistryNetworks()
	agentSettings := registry.NewAgentSettings(agentID, VMCID(cid).String(), agentNetworks, registry.EnvSettings(env), cv.agentOptions)
	if cloudProps.EphemeralDiskSize > 0 {
		agentSettings = agentSettings.AttachEphemeralDisk(registry.DefaultEphemeralDisk)
	}
	if err = cv.registryClient.Update(VMCID(cid).String(), agentSettings); err != nil {
		return "", bosherr.WrapErrorf(err, "Creating VM")
	}

	return VMCID(cid).String(), nil
}

func (cv CreateVM) updateCloudProperties(cloudProps *VMCloudProperties) {
	if cloudProps.DeployedByBoshCLI {
		cloudProps.VmNamePrefix = updateHostNameInCloudProps(cloudProps, "")
	} else {
		cloudProps.VmNamePrefix = updateHostNameInCloudProps(cloudProps, cv.timeStampForTime(time.Now().UTC()))
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

func updateHostNameInCloudProps(cloudProps *VMCloudProperties, timeStampPostfix string) string {
	if len(timeStampPostfix) == 0 {
		return cloudProps.VmNamePrefix
	} else {
		return cloudProps.VmNamePrefix + "." + timeStampPostfix
	}
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

	return &datatypes.Virtual_Guest{
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
}

func (cv CreateVM) createUserDataForInstance(agentID string, registryOptions registry.ClientOptions) (string, error) {
	serverName := fmt.Sprintf("vm-%s", agentID)
	userDataContents := registry.UserDataContentsType{
		Registry: registry.RegistryType{
			Endpoint: fmt.Sprintf("http://%s:%s@%s:%d",
				registryOptions.Username,
				registryOptions.Password,
				registryOptions.Host,
				registryOptions.Port),
		},
		Server: registry.ServerType{
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

		networkSpace, err := cv.getNetworkSpace(nw.CloudProperties.VlanID)
		if err != nil {
			return 0, 0, bosherr.WrapErrorf(err, "Network: %q, vlan id: %d", name, nw.CloudProperties.VlanID)
		}

		switch networkSpace {
		case "PRIVATE":
			if privateVlanID == 0 {
				privateVlanID = nw.CloudProperties.VlanID
			} else if privateVlanID != nw.CloudProperties.VlanID {
				return 0, 0, bosherr.Error("Only one private VLAN is supported")
			}
		case "PUBLIC":
			if publicVlanID == 0 {
				publicVlanID = nw.CloudProperties.VlanID
			} else if publicVlanID != nw.CloudProperties.VlanID {
				return 0, 0, bosherr.Error("Only one public VLAN is supported")
			}
		default:
			return 0, 0, bosherr.Errorf("Vlan id %d: unknown network type '%s'", nw.CloudProperties.VlanID, networkSpace)
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
