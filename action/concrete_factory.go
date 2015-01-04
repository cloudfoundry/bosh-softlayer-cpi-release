package action

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcbm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(softLayerClient sl.Client, options ConcreteFactoryOptions, logger boshlog.Logger) concreteFactory {
	stemcellFinder := bslcstem.NewSoftLayerFinder(softLayerClient, logger)

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(softLayerClient, logger)

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

	bmCreator := bslcbm.NewBaremetalCreator(softLayerClient, logger)
	bmFinder := bslcbm.NewBaremetalFinder(softLayerClient, logger)

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
			"create_disk": NewCreateDisk(nil),
			"delete_disk": NewDeleteDisk(nil),
			"attach_disk": NewAttachDisk(vmFinder, nil),
			"detach_disk": NewDetachDisk(vmFinder, nil),

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
