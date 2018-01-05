package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-softlayer-cpi/config"
	"bosh-softlayer-cpi/softlayer/client"
	"bosh-softlayer-cpi/softlayer/disk_service"
	"bosh-softlayer-cpi/softlayer/snapshot_service"
	"bosh-softlayer-cpi/softlayer/virtual_guest_service"

	"bosh-softlayer-cpi/logger"
	"bosh-softlayer-cpi/registry"
	"bosh-softlayer-cpi/softlayer/stemcell_service"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(
	softlayerClient client.Client,
	uuidGen boshuuid.Generator,
	cfg config.Config,
	logger logger.Logger,
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
		uuidGen,
		logger,
	)

	vmService := instance.NewSoftLayerVirtualGuestService(
		softlayerClient,
		uuidGen,
		logger,
	)

	snapshotService := snapshot.NewSoftlayerSnapshotService(
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
			"delete_vm":          NewDeleteVM(vmService, registryClient, cfg.Cloud.Properties.SoftLayer),
			"has_vm":             NewHasVM(vmService),
			"reboot_vm":          NewRebootVM(vmService),
			"set_vm_metadata":    NewSetVMMetadata(vmService),
			"configure_networks": NewConfigureNetworks(vmService, registryClient),

			// Disk management
			"has_disk":          NewHasDisk(diskService),
			"create_disk":       NewCreateDisk(diskService, vmService),
			"delete_disk":       NewDeleteDisk(diskService),
			"attach_disk":       NewAttachDisk(diskService, vmService, registryClient),
			"detach_disk":       NewDetachDisk(vmService, registryClient),
			"get_disks":         NewGetDisks(vmService),
			"set_disk_metadata": NewSetDiskMetadata(diskService),

			// Snapshot management
			"snapshot_disk":   NewSnapshotDisk(snapshotService, diskService),
			"delete_snapshot": NewDeleteSnapshot(snapshotService),

			// Others:
			"info": NewInfo(),
			"ping": NewPing(),

			// Not implemented (others):
			//   current_vm_id
			//   calculate_vm_cloud_properties
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
