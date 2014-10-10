package vm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	gomega "github.com/onsi/gomega"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

var (
	TIMEOUT          time.Duration
	POLLING_INTERVAL time.Duration
)

type SoftLayerCreator struct {
	softLayerClient        sl.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	agentOptions AgentOptions
	logger       boshlog.Logger
}

func NewSoftLayerCreator(softLayerClient sl.Client, agentEnvServiceFactory AgentEnvServiceFactory, agentOptions AgentOptions, logger boshlog.Logger) SoftLayerCreator {
	TIMEOUT = 10 * time.Minute
	POLLING_INTERVAL = 10 * time.Second

	return SoftLayerCreator{
		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,
		agentOptions:           agentOptions,
		logger:                 logger,
	}
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, cloudProps VMCloudProperties, networks Networks, env Environment) (VM, error) {
	virtualGuestTemplate := sldatatypes.SoftLayer_Virtual_Guest_Template{
		Hostname:  agentID,
		Domain:    "softlayer.com",
		StartCpus: cloudProps.StartCpus,
		MaxMemory: cloudProps.MaxMemory,
		Datacenter: sldatatypes.Datacenter{
			Name: cloudProps.Datacenter.Name,
		},
		SshKeys:                      cloudProps.SshKeys,
		HourlyBillingFlag:            true,
		LocalDiskFlag:                true,
		OperatingSystemReferenceCode: "UBUNTU_LATEST",
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	virtualGuest, err := virtualGuestService.CreateObject(virtualGuestTemplate)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating VirtualGuest from SoftLayer client")
	}

	agentEnv := NewAgentEnvForVM(agentID, strconv.Itoa(virtualGuest.Id), networks, env, c.agentOptions)

	err = c.configureMetadataOnVirtualGuest(virtualGuest.Id, agentEnv)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuest.Id))
	}

	agentEnvService := c.agentEnvServiceFactory.New()

	err = agentEnvService.Update(agentEnv)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Updating VM agent env")
	}

	vm := NewSoftLayerVM(virtualGuest.Id, c.softLayerClient, agentEnvService, c.logger)

	return vm, nil
}

func (c SoftLayerCreator) configureMetadataOnVirtualGuest(virtualGuestId int, agentEnv AgentEnv) error {
	err := c.waitForVirtualGuest(virtualGuestId, "RUNNING", TIMEOUT, POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d`", virtualGuestId))
	}

	err = c.waitForVirtualGuestToHaveNoRunningTransactions(virtualGuestId, TIMEOUT, POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d` to have no pending transactions", virtualGuestId))
	}

	err = c.setMetadataOnVirtualGuest(virtualGuestId, agentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Setting metadata on VirtualGuest `%d`", virtualGuestId))
	}

	err = c.configureMetadataDiskOnVirtualGuest(virtualGuestId)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata disk on VirtualGuest `%d`", virtualGuestId))
	}

	err = c.waitForVirtualGuest(virtualGuestId, "RUNNING", TIMEOUT, POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d`", virtualGuestId))
	}

	return nil
}

func (c SoftLayerCreator) waitForVirtualGuestToHaveNoRunningTransactions(virtualGuestId int, timeout, pollingInterval time.Duration) error {
	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	gomega.Eventually(func() int {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return len(activeTransactions)
	}, timeout, pollingInterval).Should(gomega.Equal(0), "failed waiting for virtual guest to have no active transactions")

	return nil
}

func (c SoftLayerCreator) waitForVirtualGuest(virtualGuestId int, targetState string, timeout, pollingInterval time.Duration) error {
	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	gomega.Eventually(func() string {
		vgPowerState, err := virtualGuestService.GetPowerState(virtualGuestId)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return vgPowerState.KeyName
	}, timeout, pollingInterval).Should(gomega.Equal(targetState), fmt.Sprintf("failed waiting for virtual guest to be %s", targetState))

	return nil
}

func (c SoftLayerCreator) setMetadataOnVirtualGuest(virtualGuestId int, agentEnv AgentEnv) error {
	metadata, err := json.Marshal(agentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent environment metadata")
	}

	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	success, err := virtualGuestService.SetMetadata(virtualGuestId, string(metadata))
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Setting metadata on VirtualGuest `%d`", virtualGuestId))
	}

	if !success {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to set metadata on VirtualGuest `%d`", virtualGuestId))
	}

	return nil
}

func (c SoftLayerCreator) configureMetadataDiskOnVirtualGuest(virtualGuestId int) error {
	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	_, err = virtualGuestService.ConfigureMetadataDisk(virtualGuestId)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuestId))
	}

	return nil
}

func (c SoftLayerCreator) resolveNetworkIP(networks Networks) (string, error) {
	var network Network

	switch len(networks) {
	case 0:
		return "", bosherr.New("Expected exactly one network; received zero")
	case 1:
		network = networks.First()
	default:
		return "", bosherr.New("Expected exactly one network; received multiple")
	}

	if network.IsDynamic() {
		return "", nil
	}

	return network.IP, nil
}
