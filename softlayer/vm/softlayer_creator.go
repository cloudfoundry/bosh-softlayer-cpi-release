package vm

import (
	"bytes"
	"fmt"
	"time"
	"net"
	"net/url"
	"encoding/json"
	"text/template"

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
	bslcommon.POLLING_INTERVAL = 20 * time.Second
	bslcommon.MAX_RETRY_COUNT = 5

	return SoftLayerCreator{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
	}
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestTemplate, err := CreateVirtualGuestTemplate(agentID, stemcell, cloudProps, networks, env, c.agentOptions)
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
		err = bslcommon.WaitForVirtualGuest(c.softLayerClient, virtualGuest.Id, "RUNNING", bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id `%d`", virtualGuest.Id))
		}
	} else {
		err = bslcommon.AttachEphemeralDiskToVirtualGuest(c.softLayerClient, virtualGuest.Id, cloudProps.EphemeralDiskSize, bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Attaching ephemeral disk to VirtualGuest `%d`", virtualGuest.Id))
		}
	}

	agentEnvService := c.agentEnvServiceFactory.New(virtualGuest.Id)
	vm := NewSoftLayerVM(virtualGuest.Id, c.softLayerClient, util.GetSshClient(), agentEnvService, c.logger, TIMEOUT_TRANSACTIONS_DELETE_VM)


	if len(cloudProps.BoshIp) = 0{
		virtualGuest, err = bslcommon.GetObjectDetailsOnVirtualGuest(vm.softLayerClient, virtualGuest.Id)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapErrorf(err, "Can not get details from virtual guest with id: %d.", virtualGuest.Id)
		}
		c.updateEtcHostsOfBoshInit(fmt.Sprintf("%s  %s", virtualGuest.PrimaryBackendIpAddress, virtualGuest.FullyQualifiedDomainName))

		// Update mbus url setting for bosh director
		metadata, err := bslcommon.GetUserMetadataOnVirtualGuest(vm.softLayerClient, virtualGuest.Id)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapErrorf(err, "Can not get metadata from virtual guest with id: %d.", virtualGuest.Id)
		}
		agentEnv, err := NewAgentEnvFromJSON(metadata)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapErrorf(err, "Can not unmarshal metadata from virutal guest with id: %d.", virtualGuest.Id)
		}

		//Construct mbus url with new director ip
		mbus,err := c.parseMbusURL(c.agentOptions.Mbus, virtualGuest.PrimaryBackendIpAddress)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapErrorf(err, "Can not construct mbus url.")
		}
		agentEnv.Mbus = mbus

		metadata, err = json.Marshal(agentEnv)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, "Marshalling agent environment metadata")
		}

		err = bslcommon.ConfigureMetadataOnVirtualGuest(vm.softLayerClient, virtualGuest.Id, string(metadata), bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
		if err != nil {
			return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuest.Id))
		}
	}

	return vm, nil
}

// Private methods

func (c SoftLayerCreator) resolveNetworkIP(networks Networks) (string, error) {
	var network Network

	switch len(networks) {
	case 0:
		return "", bosherr.Error("Expected exactly one network; received zero")
	case 1:
		network = networks.First()
	default:
		return "", bosherr.Error("Expected exactly one network; received multiple")
	}

	if network.IsDynamic() {
		return "", nil
	}

	return network.IP, nil
}

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
	} else {
		return fmt.Sprintf("https://%s:%s", primaryBackendIpAddress, port), nil
	}
	return "", nil
}

func (c SoftLayerCreator) updateEtcHostsOfBoshInit(record string) (err error) {
	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("etc-hosts").Parse(etcHostsTemplate))

	err = t.Execute(buffer, record)
	if err != nil {
		return bosherr.WrapError(err, "Generating config from template")
	}

	err = boshsys.FileSystem.WriteFile("/etc/hosts", buffer.Bytes())
	if err != nil {
		return bosherr.WrapError(err, "Writing to /etc/hosts")
	}
	return nil
}

const etcHostsTemplate = `127.0.0.1 localhost
{{.}}
`

