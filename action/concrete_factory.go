package action

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcdsk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(softLayerClient sl.Client, options ConcreteFactoryOptions, logger boshlog.Logger) concreteFactory {
	stemcellFinder := bslcstem.NewSoftLayerFinder(softLayerClient, logger)

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(logger)

	vmCreator := bslcvm.NewSoftLayerCreator(
		softLayerClient,
		agentEnvServiceFactory,
		options.Agent,
		logger,
	)

	vmFinder := bslcvm.NewSoftLayerFinder(
		softLayerClient,
		agentEnvServiceFactory,
		logger,
	)

	diskCreator := bslcdsk.NewSoftLayerCreator(
		softLayerClient,
		logger,
	)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(stemcellFinder),
			"delete_stemcell": NewDeleteStemcell(stemcellFinder),

			// VM management
			"create_vm":          NewCreateVM(stemcellFinder, vmCreator),
			"delete_vm":          NewDeleteVM(vmFinder),
			"has_vm":             NewHasVM(vmFinder),
			"reboot_vm":          NewRebootVM(vmFinder),
			"set_vm_metadata":    NewSetVMMetadata(vmFinder),
			"configure_networks": NewConfigureNetworks(vmFinder),

			// Disk management
			"create_disk": NewCreateDisk(diskCreator),
			"delete_disk": NewDeleteDisk(nil),
			"attach_disk": NewAttachDisk(vmFinder, nil),
			"detach_disk": NewDetachDisk(vmFinder, nil),

			// Not implemented:
			//   current_vm_id
			//   snapshot_disk
			//   delete_snapshot
			//   get_disks
			//   ping
		},
	}
}

func (f concreteFactory) Create(method string) (Action, error) {
	action, found := f.availableActions[method]
	if !found {
		return nil, bosherr.New("Could not create action with method %s", method)
	}

	return action, nil
}
