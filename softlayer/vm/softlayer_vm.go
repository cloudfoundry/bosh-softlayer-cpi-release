package vm

import (
	"encoding/json"
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
)

const softLayerVMtag = "SoftLayerVM"

type SoftLayerVM struct {
	id int

	softLayerClient sl.Client
	agentEnvService AgentEnvService

	logger boshlog.Logger
}

func NewSoftLayerVM(id int, softLayerClient sl.Client, agentEnvService AgentEnvService, logger boshlog.Logger) SoftLayerVM {
	bslcommon.TIMEOUT = 10 * time.Minute
	bslcommon.POLLING_INTERVAL = 10 * time.Second

	return SoftLayerVM{
		id: id,

		softLayerClient: softLayerClient,
		agentEnvService: agentEnvService,

		logger: logger,
	}
}

func (vm SoftLayerVM) ID() int { return vm.id }

func (vm SoftLayerVM) Delete() error {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	deleted, err := virtualGuestService.DeleteObject(vm.ID())
	if err != nil {
		return bosherr.WrapError(err, "Deleting SoftLayer VirtualGuest from client")
	}

	if !deleted {
		return bosherr.WrapError(nil, "Did not delete SoftLayer VirtualGuest from client")
	}

	return nil
}

func (vm SoftLayerVM) Reboot() error {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	rebooted, err := virtualGuestService.RebootSoft(vm.ID())
	if err != nil {
		return bosherr.WrapError(err, "Rebooting (soft) SoftLayer VirtualGuest from client")
	}

	if !rebooted {
		return bosherr.WrapError(nil, "Did not reboot (soft) SoftLayer VirtualGuest from client")
	}

	return nil
}

func (vm SoftLayerVM) SetMetadata(vmMetadata VMMetadata) error {
	//TODO: reconcile this metadata (from VMMetadata) abd SoftLayerCreator.Run metadata
	metadata, err := json.Marshal(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling VM metadata")
	}

	err = bslcommon.ConfigureMetadataOnVirtualGuest(vm.softLayerClient, vm.id, string(metadata), bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", vm.id))
	}

	return nil
}

func (vm SoftLayerVM) ConfigureNetworks(networks Networks) error {
	vm.logger.Info(softLayerVMtag, "Not yet implemented!")

	return nil
}

func (vm SoftLayerVM) AttachDisk(disk bslcdisk.Disk) error {
	vm.logger.Info(softLayerVMtag, "Not yet implemented!")

	return nil
}

func (vm SoftLayerVM) DetachDisk(disk bslcdisk.Disk) error {
	vm.logger.Info(softLayerVMtag, "Not yet implemented!")

	return nil
}
