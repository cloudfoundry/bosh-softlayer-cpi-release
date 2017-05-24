package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/softlayer/client"
	"bosh-softlayer-cpi/softlayer/disk_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	"bosh-softlayer-cpi/registry"
	"bosh-softlayer-cpi/softlayer/stemcell_service"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(
	softlayerClient client.Client,
	cfg config.Config,
	logger boshlog.Logger,
) concreteFactory {

	registryClient := registry.NewHTTPClient(
		cfg.Cloud.Properties.Registry,
		logger,
	)

	diskService := disk.NewSoftlayerDiskService(
		softlayerClient,
		logger,
	)

	stemcellService := stemcell.NewSoftlayerStemcellService(
		softlayerClient,
		logger,
	)

	vmService := instance.NewSoftLayerVirtualGuestService(
		softlayerClient,
		logger,
	)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(stemcellService),
			"delete_stemcell": NewDeleteStemcell(stemcellService),

			// VM management
			"create_vm": NewCreateVM(
				stemcellService,
				vmService,
				registryClient,
				cfg.Cloud.Properties.Registry,
				cfg.Cloud.Properties.Agent,
				cfg.Cloud.Properties.SoftLayer,
			),
			"delete_vm":          NewDeleteVM(vmService, registryClient),
			"has_vm":             NewHasVM(vmService),
			"reboot_vm":          NewRebootVM(vmService),
			"set_vm_metadata":    NewSetVMMetadata(vmService),
			"configure_networks": NewConfigureNetworks(vmService, registryClient),

			// Disk management
			"has_disk":    NewHasDisk(diskService),
			"create_disk": NewCreateDisk(diskService, vmService),
			"delete_disk": NewDeleteDisk(diskService),
			"attach_disk": NewAttachDisk(diskService, vmService, registryClient),
			"detach_disk": NewDetachDisk(vmService, registryClient),
			"get_disks":   NewGetDisks(vmService),

			// Others:
			"ping": NewPing(),

			// Not implemented (disk related):
			//   snapshot_disk
			//   delete_snapshot

			// Not implemented (others):
			//   current_vm_id
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
