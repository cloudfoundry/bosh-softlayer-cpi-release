package common

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"text/template"
	"time"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

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
	return now.Format("20060102-030405-") + strconv.Itoa(int(now.UnixNano()/1e6-now.Unix()*1e3))
}

func CreateVirtualGuestTemplate(stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks) (sldatatypes.SoftLayer_Virtual_Guest_Template, error) {
	for _, network := range networks {
		switch network.Type {
		case "dynamic":
			if value, ok := network.CloudProperties["PrimaryNetworkComponent"]; ok {
				networkComponent := value.(map[string]interface{})
				if value1, ok := networkComponent["NetworkVlan"]; ok {
					networkValn := value1.(map[string]interface{})
					if value2, ok := networkValn["Id"]; ok {
						cloudProps.PrimaryNetworkComponent = sldatatypes.PrimaryNetworkComponent{
							NetworkVlan: sldatatypes.NetworkVlan{
								Id: int(value2.(float64)),
							},
						}
					}
				}
			}
			if value, ok := network.CloudProperties["PrimaryBackendNetworkComponent"]; ok {
				networkComponent := value.(map[string]interface{})
				if value1, ok := networkComponent["NetworkVlan"]; ok {
					networkValn := value1.(map[string]interface{})
					if value2, ok := networkValn["Id"]; ok {
						cloudProps.PrimaryBackendNetworkComponent = sldatatypes.PrimaryBackendNetworkComponent{
							NetworkVlan: sldatatypes.NetworkVlan{
								Id: int(value2.(float64)),
							},
						}
					}
				}
			}
			if value, ok := network.CloudProperties["PrivateNetworkOnlyFlag"]; ok {
				privateOnly := value.(bool)
				cloudProps.PrivateNetworkOnlyFlag = privateOnly
			}
		default:
			continue
		}
	}

	virtualGuestTemplate := sldatatypes.SoftLayer_Virtual_Guest_Template{
		Hostname:  cloudProps.VmNamePrefix,
		Domain:    cloudProps.Domain,
		StartCpus: cloudProps.StartCpus,
		MaxMemory: cloudProps.MaxMemory,

		Datacenter: sldatatypes.Datacenter{
			Name: cloudProps.Datacenter.Name,
		},

		BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
			GlobalIdentifier: stemcell.Uuid(),
		},

		SshKeys: cloudProps.SshKeys,

		HourlyBillingFlag: cloudProps.HourlyBillingFlag,
		LocalDiskFlag:     cloudProps.LocalDiskFlag,

		DedicatedAccountHostOnlyFlag:   cloudProps.DedicatedAccountHostOnlyFlag,
		BlockDevices:                   cloudProps.BlockDevices,
		NetworkComponents:              cloudProps.NetworkComponents,
		PrivateNetworkOnlyFlag:         cloudProps.PrivateNetworkOnlyFlag,
		PrimaryNetworkComponent:        &cloudProps.PrimaryNetworkComponent,
		PrimaryBackendNetworkComponent: &cloudProps.PrimaryBackendNetworkComponent,
	}

	return virtualGuestTemplate, nil
}

func CreateAgentUserData(agentID string, cloudProps VMCloudProperties, networks Networks, env Environment, agentOptions AgentOptions) AgentEnv {
	agentName := fmt.Sprintf("vm-%s", agentID)
	disks := CreateDisksSpec(cloudProps.EphemeralDiskSize)
	agentEnv := NewAgentEnvForVM(agentID, agentName, networks, disks, env, agentOptions)
	return agentEnv
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

func UpdateEtcHostsOfBoshInit(fileName string, record string) (err error) {
	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("etc-hosts").Parse(ETC_HOSTS_TEMPLATE))

	err = t.Execute(buffer, record)
	if err != nil {
		return bosherr.WrapError(err, "Generating config from template")
	}

	logger := boshlog.NewWriterLogger(boshlog.LevelError, os.Stderr, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)

	err = fs.WriteFile(fileName, buffer.Bytes())
	if err != nil {
		return bosherr.WrapError(err, "Writing to /etc/hosts")
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

const ETC_HOSTS_TEMPLATE = `127.0.0.1 localhost
{{.}}
`
