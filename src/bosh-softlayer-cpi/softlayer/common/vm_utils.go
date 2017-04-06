package common

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"bufio"

	"encoding/json"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	sl "github.com/maximilien/softlayer-go/softlayer"
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

func CreateVirtualGuestTemplate(stemcellUuid string, cloudProps VMCloudProperties, userData string, publicVlanId int, privateVlanId int) (sldatatypes.SoftLayer_Virtual_Guest_Template, error) {
	var publicNetworkComponent *sldatatypes.PrimaryNetworkComponent
	var privateNetworkComponent *sldatatypes.PrimaryBackendNetworkComponent

	if publicVlanId != 0 {
		publicNetworkComponent = &sldatatypes.PrimaryNetworkComponent{
			NetworkVlan: sldatatypes.NetworkVlan{
				Id: publicVlanId,
			},
		}
	}

	privateNetworkComponent = &sldatatypes.PrimaryBackendNetworkComponent{
		NetworkVlan: sldatatypes.NetworkVlan{
			Id: privateVlanId,
		},
	}

	virtualGuestTemplate := sldatatypes.SoftLayer_Virtual_Guest_Template{
		Hostname:  cloudProps.VmNamePrefix,
		Domain:    cloudProps.Domain,
		StartCpus: cloudProps.StartCpus,
		MaxMemory: cloudProps.MaxMemory,

		Datacenter: sldatatypes.Datacenter{
			Name: cloudProps.Datacenter,
		},

		BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
			GlobalIdentifier: stemcellUuid,
		},

		SshKeys: cloudProps.SshKeys,

		HourlyBillingFlag: cloudProps.HourlyBillingFlag,
		LocalDiskFlag:     cloudProps.LocalDiskFlag,

		DedicatedAccountHostOnlyFlag:   cloudProps.DedicatedAccountHostOnlyFlag,
		NetworkComponents:              cloudProps.NetworkComponents,
		PrivateNetworkOnlyFlag:         publicNetworkComponent == nil,
		PrimaryNetworkComponent:        publicNetworkComponent,
		PrimaryBackendNetworkComponent: privateNetworkComponent,

		UserData: []sldatatypes.UserData{
			{
				Value: userData,
			},
		},
	}

	return virtualGuestTemplate, nil
}

func CreateAgentUserData(agentID string, cloudProps VMCloudProperties, networks Networks, env Environment, agentOptions AgentOptions) AgentEnv {
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

func UpdateDeviceName(vmID int, virtualGuestService sl.SoftLayer_Virtual_Guest_Service, cloudProps VMCloudProperties) (err error) {
	deviceName := sldatatypes.SoftLayer_Virtual_Guest{
		Hostname: cloudProps.VmNamePrefix,
		Domain:   cloudProps.Domain,
		FullyQualifiedDomainName: cloudProps.VmNamePrefix + "." + cloudProps.Domain,
	}

	_, err = virtualGuestService.EditObject(vmID, deviceName)
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
