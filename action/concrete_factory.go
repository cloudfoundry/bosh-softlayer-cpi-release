package action

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshcmd "bosh/platform/commands"
	boshsys "bosh/system"
	boshuuid "bosh/uuid"

	bslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcutil "github.com/maximilien/bosh-softlayer-cpi/util"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(
	SoftLayerClient bslcpi.Client,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
	compressor boshcmd.Compressor,
	sleeper bslcutil.Sleeper,
	options ConcreteFactoryOptions,
	logger boshlog.Logger,
) concreteFactory {
	stemcellImporter := bslcstem.NewFSImporter(
		options.StemcellsDir,
		fs,
		uuidGen,
		compressor,
		logger,
	)

	stemcellFinder := bslcstem.NewFSFinder(options.StemcellsDir, fs, logger)

	hostBindMounts := bslcvm.NewFSHostBindMounts(
		options.HostEphemeralBindMountsDir,
		options.HostPersistentBindMountsDir,
		sleeper,
		fs,
		cmdRunner,
		logger,
	)

	guestBindMounts := bslcvm.NewFSGuestBindMounts(
		options.GuestEphemeralBindMountPath,
		options.GuestPersistentBindMountsDir,
		logger,
	)

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(logger)

	vmCreator := bslcvm.NewSoftLayerCreator(
		uuidGen,
		SoftLayerClient,
		agentEnvServiceFactory,
		hostBindMounts,
		guestBindMounts,
		options.Agent,
		logger,
	)

	vmFinder := bslcvm.NewSoftLayerFinder(
		SoftLayerClient,
		agentEnvServiceFactory,
		hostBindMounts,
		guestBindMounts,
		logger,
	)

	diskCreator := bslcdisk.NewFSCreator(
		options.DisksDir,
		fs,
		uuidGen,
		cmdRunner,
		logger,
	)

	diskFinder := bslcdisk.NewFSFinder(options.DisksDir, fs, logger)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(stemcellImporter),
			"delete_stemcell": NewDeleteStemcell(stemcellFinder),

			// VM management
			"create_vm":          NewCreateVM(stemcellFinder, vmCreator),
			"delete_vm":          NewDeleteVM(vmFinder),
			"has_vm":             NewHasVM(vmFinder),
			"reboot_vm":          NewRebootVM(),
			"set_vm_metadata":    NewSetVMMetadata(),
			"configure_networks": NewConfigureNetworks(),

			// Disk management
			"create_disk": NewCreateDisk(diskCreator),
			"delete_disk": NewDeleteDisk(diskFinder),
			"attach_disk": NewAttachDisk(vmFinder, diskFinder),
			"detach_disk": NewDetachDisk(vmFinder, diskFinder),

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
