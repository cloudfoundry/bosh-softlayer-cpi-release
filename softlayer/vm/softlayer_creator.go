package vm

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"text/template"
	"time"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshsys "github.com/cloudfoundry/bosh-utils/system"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

type SoftLayerCreator struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
}

func NewSoftLayerCreator(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, agentOptions AgentOptions, logger boshlog.Logger) SoftLayerCreator {
	bslcommon.TIMEOUT = 60 * time.Minute
	bslcommon.POLLING_INTERVAL = 10 * time.Second

	return SoftLayerCreator{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
	}
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestTemplate, err := CreateVirtualGuestTemplate(stemcell, cloudProps)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating virtual guest template")
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(virtualGuestTemplate)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuest from SoftLayer client")
	}

	if cloudProps.EphemeralDiskSize == 0 {
		err = bslcommon.WaitForVirtualGuest(c.softLayerClient, virtualGuest.Id, "RUNNING")
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id `%d`", virtualGuest.Id))
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, c.logger)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
		}
	}

	err = bslcommon.WaitForVirtualGuestLastCompleteTransaction(c.softLayerClient, virtualGuest.Id, "Service Setup")
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", virtualGuest.Id)
	}

	virtualGuest, err = bslcommon.GetObjectDetailsOnVirtualGuest(c.softLayerClient, virtualGuest.Id)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapErrorf(err, "Cannot get details from virtual guest with id: %d.", virtualGuest.Id)
	}

	softlayerFileService := NewSoftlayerFileService(util.GetSshClient(), virtualGuest, c.logger)
	agentEnvService := c.agentEnvServiceFactory.New(softlayerFileService, strconv.Itoa(virtualGuest.Id))

	agentEnv := CreateAgentUserData(agentID, cloudProps, networks, env, c.agentOptions)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapErrorf(err, "Cannot agent env for virtual guest with id: %d.", virtualGuest.Id)
	}

	if len(cloudProps.BoshIp) == 0 {
		// update /etc/hosts file of bosh-init vm
		c.updateEtcHostsOfBoshInit(fmt.Sprintf("%s  %s", virtualGuest.PrimaryBackendIpAddress, virtualGuest.FullyQualifiedDomainName))
		// Update mbus url setting for bosh director: construct mbus url with new director ip
		mbus, err := c.parseMbusURL(c.agentOptions.Mbus, virtualGuest.PrimaryBackendIpAddress)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapErrorf(err, "Cannot construct mbus url.")
		}
		agentEnv.Mbus = mbus
	}

	err = agentEnvService.Update(agentEnv)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Updating VM's agent env")
	}

	vm := NewSoftLayerVM(virtualGuest.Id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger)

	return vm, nil
}

// Private methods
func (c SoftLayerCreator) parseMbusURL(mbusURL string, primaryBackendIpAddress string) (string, error) {
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
		return fmt.Sprintf("https://%s:%s@%s:%s", username, password, primaryBackendIpAddress, port), nil
	}

	return fmt.Sprintf("https://%s:%s", primaryBackendIpAddress, port), nil
}

func (c SoftLayerCreator) updateEtcHostsOfBoshInit(record string) (err error) {
	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("etc-hosts").Parse(ETC_HOSTS_TEMPLATE))

	err = t.Execute(buffer, record)
	if err != nil {
		return bosherr.WrapError(err, "Generating config from template")
	}

	fileSystem := boshsys.NewOsFileSystemWithStrictTempRoot(c.logger)
	err = fileSystem.WriteFile("/etc/hosts", buffer.Bytes())
	if err != nil {
		return bosherr.WrapError(err, "Writing to /etc/hosts")
	}

	return nil
}

const ETC_HOSTS_TEMPLATE = `127.0.0.1 localhost
{{.}}
`
