package common

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	datatypes "github.com/softlayer/softlayer-go/datatypes"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"bufio"
	"encoding/json"

	bslc "bosh-softlayer-cpi/softlayer/client"
	snet "bosh-softlayer-cpi/softlayer/networks"
	"github.com/softlayer/softlayer-go/sl"
)

func CreateDisksSpec(ephemeralDiskSize int) DisksSpec {
	disks := DisksSpec{}
	if ephemeralDiskSize > 0 {
		disks = DisksSpec{
			Ephemeral:  "/dev/xvdc",
			Persistent: nil,
		}
	}

	return disks
}

func TimeStampForTime(now time.Time) string {
	//utilize the constants list in the http://golang.org/src/time/format.go file to get the expect time formats
	return now.Format("20060102-030405-") + fmt.Sprintf("%03d", int(now.UnixNano()/1e6-now.Unix()*1e3))
}

func CreateVirtualGuestTemplate(stemcellUuid string, cloudProps VMCloudProperties, userData string, publicVlanId int, privateVlanId int) *datatypes.Virtual_Guest {
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

func CreateAgentUserData(agentID string, cloudProps VMCloudProperties, networks snet.Networks, env Environment, agentOptions AgentOptions) AgentEnv {
	agentName := fmt.Sprintf("vm-%s", agentID)
	disks := CreateDisksSpec(cloudProps.EphemeralDiskSize)
	agentEnv := NewAgentEnvForVM(agentID, agentName, networks, disks, env, agentOptions)
	return agentEnv
}

func CreateUserDataForInstance(agentID string, registryOptions RegistryOptions) (string, error) {
	serverName := fmt.Sprintf("vm-%s", agentID)
	userDataContents := UserDataContentsType{
		Registry: RegistryType{
			Endpoint: fmt.Sprintf("http://%s:%s@%s:%d",
				registryOptions.Username,
				registryOptions.Password,
				registryOptions.Host,
				registryOptions.Port),
		},
		Server: ServerType{
			Name: serverName,
		},
	}
	contentsBytes, err := json.Marshal(userDataContents)
	if err != nil {
		return "", bosherr.WrapError(err, "Preparing user data contents")
	}

	return base64.RawURLEncoding.EncodeToString(contentsBytes), nil
}

func UpdateDavConfig(config *DavConfig, directorIP string) (err error) {
	url := (*config)["endpoint"].(string)
	mbus, err := ParseMbusURL(url, directorIP)
	if err != nil {
		return bosherr.WrapError(err, "Parsing Mbus URL")
	}

	(*config)["endpoint"] = mbus

	return nil
}

func ParseMbusURL(mbusURL string, primaryBackendIpAddress string) (string, error) {
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

func UpdateEtcHostsOfBoshInit(path string, record string) (err error) {
	logger := boshlog.NewWriterLogger(boshlog.LevelError, os.Stderr, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)

	if !fs.FileExists(path) {
		err := fs.WriteFile(path, []byte{})
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating the new file %s if it does not exist", path)
		}
	}

	fileHandle, err := fs.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening file %s", path)
	}

	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	length, err := fmt.Fprintln(writer, "\n"+record)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing '%s' to Writer", record)
	}
	if length != len(record)+2 {
		return bosherr.Errorf("The number (%d) of bytes written in Writer is not equal to the length (%d) of string", length, len(record)+2)
	}

	err = writer.Flush()
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing '%s' to file %s", record, path)
	}

	return nil
}

func UpdateDeviceName(vmID int, virtualGuestService bslc.Client, cloudProps VMCloudProperties) (err error) {
	editingVirtualGuest := &datatypes.Virtual_Guest{
		Hostname: sl.String(cloudProps.VmNamePrefix),
		Domain:   sl.String(cloudProps.Domain),
		FullyQualifiedDomainName: sl.String(cloudProps.VmNamePrefix + "." + cloudProps.Domain),
	}

	_, err = virtualGuestService.EditInstance(editingVirtualGuest)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to update properties for virtualGuest with id: %d", vmID)
	}
	return nil
}

func GetLocalIPAddressOfGivenInterface(networkInterface string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to get network interfaces")
	}

	for _, i := range interfaces {
		if i.Name == networkInterface {
			addrs, err := i.Addrs()
			if err != nil {
				return "", bosherr.WrapErrorf(err, fmt.Sprintf("Failed to get interface addresses for %s", networkInterface))
			}
			for _, addr := range addrs {
				ipnet, _ := addr.(*net.IPNet)
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", bosherr.Error(fmt.Sprintf("Failed to get IP address of %s", networkInterface))
}

func GetVlanIds(softLayerClient bslc.Client, networks snet.Networks) (int, int, error) {
	var publicVlanID, privateVlanID int

	for name, nw := range networks {
		networkSpace, err := getNetworkSpace(softLayerClient, nw.CloudProperties.VlanID)
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

func getNetworkSpace(softLayerClient bslc.Client, vlanID int) (string, error) {
	networkVlan, err := softLayerClient.GetVlan(vlanID, bslc.NETWORK_DEFAULT_VLAN)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Getting vlan info with id `%d`", vlanID)
	}
	return *networkVlan.NetworkSpace, nil
}
