package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcbm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(softLayerClient sl.Client, options ConcreteFactoryOptions, logger boshlog.Logger, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem) concreteFactory {

	stemcellFinder := bslcstem.NewSoftLayerFinder(softLayerClient, logger)

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(options.AgentEnvService, options.Registry, logger)

	vmCreator := bslcvm.NewSoftLayerCreator(
		softLayerClient,
		agentEnvServiceFactory,
		options.Agent,
		logger,
		uuidGenerator,
		fs,
	)

	vmFinder := bslcvm.NewSoftLayerFinder(
		softLayerClient,
		agentEnvServiceFactory,
		logger,
		uuidGenerator,
		fs,
	)

	bmCreator := bslcbm.NewBaremetalCreator(softLayerClient, logger)
	bmFinder := bslcbm.NewBaremetalFinder(softLayerClient, logger)

	diskCreator := bslcdisk.NewSoftLayerDiskCreator(
		softLayerClient,
		logger,
	)

	diskFinder := bslcdisk.NewSoftLayerDiskFinder(
		softLayerClient,
		logger,
	)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(stemcellFinder),
			"delete_stemcell": NewDeleteStemcell(stemcellFinder, logger),

			// VM management
			"create_vm":          NewCreateVM(stemcellFinder, vmCreator),
			"delete_vm":          NewDeleteVM(vmFinder),
			"has_vm":             NewHasVM(vmFinder),
			"reboot_vm":          NewRebootVM(vmFinder),
			"set_vm_metadata":    NewSetVMMetadata(vmFinder),
			"configure_networks": NewConfigureNetworks(vmFinder),

			// Disk management
			"create_disk": NewCreateDisk(diskCreator),
			"delete_disk": NewDeleteDisk(diskFinder),
			"attach_disk": NewAttachDisk(vmFinder, diskFinder),
			"detach_disk": NewDetachDisk(vmFinder, diskFinder),

			"establish_bare_metal_env": NewEstablishBareMetalEnv(bmCreator, bmFinder),

			// Not implemented (disk related):
			//   snapshot_disk
			//   delete_snapshot
			//   get_disks

			// Not implemented (others):
			//   current_vm_id
			//   ping
		},
	}
}

func (f concreteFactory) Create(method string) (Action, error) {
	action, found := f.availableActions[method]
	if !found {
		return nil, bosherr.Errorf("Could not create action with method %s", method)
	}

	return action, nil
}
