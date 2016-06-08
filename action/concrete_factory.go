package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bmsclient "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	slclient "github.com/maximilien/softlayer-go/client"

	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(options ConcreteFactoryOptions, logger boshlog.Logger) concreteFactory {
	fs := boshsys.NewOsFileSystem(logger)
	uuidGenerator := boshuuid.NewGenerator()

	softLayerClient := slclient.NewSoftLayerClient(options.Softlayer.Username, options.Softlayer.ApiKey)
	baremetalClient := bmsclient.NewBmpClient(options.Baremetal.Username, options.Baremetal.Password, options.Baremetal.EndPoint, nil, "")

	stemcellFinder := bslcstem.NewSoftLayerFinder(softLayerClient, logger)

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(options.AgentEnvService, options.Registry, logger)

	vmFinder := bslcvm.NewSoftLayerFinder(
		softLayerClient,
		baremetalClient,
		agentEnvServiceFactory,
		logger,
		uuidGenerator,
		fs,
	)

	vmCreatorProvider := NewProvider(
		softLayerClient,
		baremetalClient,
		options,
		logger,
		uuidGenerator,
		fs,
	)

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
			"create_vm":          NewCreateVM(stemcellFinder, vmCreatorProvider),
			"delete_vm":          NewDeleteVM(vmFinder),
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
