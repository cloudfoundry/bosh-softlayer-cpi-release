package action

import (
	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "bosh-softlayer-cpi/softlayer/vm"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bsl "bosh-softlayer-cpi/softlayer/client"

	. "bosh-softlayer-cpi/softlayer/common"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(options ConcreteFactoryOptions, logger boshlog.Logger) concreteFactory {
	softLayerClient := bsl.NewSoftlayerClientSession(bsl.SoftlayerAPIEndpointPublicDefault, options.Softlayer.Username, options.Softlayer.ApiKey, true, 300)
	repClientFactory := bsl.NewClientFactory(bsl.NewSoftLayerClientManager(softLayerClient))
	client := repClientFactory.CreateClient()

	stemcellFinder := bslcstem.NewSoftLayerStemcellFinder(client, logger)

	agentEnvServiceFactory := NewSoftLayerAgentEnvServiceFactory(options.Registry, logger)

	vmFinder := bslcvm.NewSoftLayerFinder(
		client,
		agentEnvServiceFactory,
		logger,
	)

	vmCreatorProvider := NewCreatorProvider(
		client,
		options,
		logger,
	)

	vmDeleterProvider := NewDeleterProvider(
		client,
		logger,
		vmFinder,
	)

	diskCreator := bslcdisk.NewSoftLayerDiskCreator(
		client,
		logger,
	)

	diskFinder := bslcdisk.NewSoftLayerDiskFinder(
		client,
		logger,
	)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(stemcellFinder),
			"delete_stemcell": NewDeleteStemcell(stemcellFinder, logger),

			// VM management
			"create_vm":          NewCreateVM(stemcellFinder, vmCreatorProvider, options),
			"delete_vm":          NewDeleteVM(vmDeleterProvider, options),
			"has_vm":             NewHasVM(vmFinder),
			"reboot_vm":          NewRebootVM(vmFinder),
			"set_vm_metadata":    NewSetVMMetadata(vmFinder),
			"configure_networks": NewConfigureNetworks(vmFinder),

			// Disk management
			"create_disk": NewCreateDisk(vmFinder, diskCreator),
			"delete_disk": NewDeleteDisk(diskFinder),
			"attach_disk": NewAttachDisk(vmFinder, diskFinder),
			"detach_disk": NewDetachDisk(vmFinder, diskFinder),

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
