package instance

import (
	"bufio"
	"fmt"
	"net"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boslc "bosh-softlayer-cpi/softlayer/client"

	"net/url"
)

func (vg SoftlayerVirtualGuestService) Create(vmProps *Properties, networks Networks, registryEndpoint string) (int, error) {
	virtualGuest, err := vg.softlayerClient.CreateInstance(&vmProps.VirtualGuestTemplate)
	if err != nil {
		return 0, bosherr.WrapError(err, "Creating virtualGuest")
	}

	virtualGuest, err = vg.softlayerClient.GetInstance(*virtualGuest.Id, boslc.INSTANCE_DETAIL_MASK)
	if err != nil {
		return 0, bosherr.WrapError(err, "Getting virtualGuest")
	}

	if vmProps.DeployedByBoshCLI {
		err := vg.updateEtcHostsOfBoshInit("/etc/hosts", fmt.Sprintf("%s  %s", *virtualGuest.PrimaryBackendIpAddress, *virtualGuest.FullyQualifiedDomainName))
		if err != nil {
			return 0, bosherr.WrapErrorf(err, "Updating BOSH director hostname/IP mapping entry in /etc/hosts")
		}
	} else {
		boshIP, err := vg.getLocalIPAddressOfGivenInterface("eth0")
		if err != nil {
			return 0, bosherr.WrapError(err, "Failed to get IP address of eth0 in local")
		}

		mbus, err := vg.parseMbusURL(vmProps.agentOption.Mbus, boshIP)
		if err != nil {
			return 0, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		vmProps.agentOption.Mbus = mbus

		switch vmProps.agentOption.Blobstore.Provider {
		case blobstoreTypeDav:
			davConf := DavConfig(vmProps.agentOption.Blobstore.Options)
			vg.updateDavConfig(&davConf, boshIP)
		}
	}

	return *virtualGuest.Id, nil
}

func (vg SoftlayerVirtualGuestService) CleanUp(id int) {
	if err := vg.Delete(id); err != nil {
		vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Failed cleaning up Softlayer VirtualGuest '%s': %v", id, err)
	}

}

func (vg SoftlayerVirtualGuestService) updateEtcHostsOfBoshInit(path string, record string) (err error) {
	fs := boshsys.NewOsFileSystem(vg.logger)

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

func (vg SoftlayerVirtualGuestService) getLocalIPAddressOfGivenInterface(networkInterface string) (string, error) {
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

func (vg SoftlayerVirtualGuestService) parseMbusURL(mbusURL string, primaryBackendIpAddress string) (string, error) {
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

func (vg SoftlayerVirtualGuestService) updateDavConfig(config *DavConfig, directorIP string) (err error) {
	url := (*config)["endpoint"].(string)
	mbus, err := vg.parseMbusURL(url, directorIP)
	if err != nil {
		return bosherr.WrapError(err, "Parsing Mbus URL")
	}

	(*config)["endpoint"] = mbus

	return nil
}
