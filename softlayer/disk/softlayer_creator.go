package disk

import (
	"fmt"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

type SoftLayerCreator struct {
	softLayerClient sl.Client
	logger          boshlog.Logger
}

func NewSoftLayerDiskCreator(client sl.Client, logger boshlog.Logger) SoftLayerCreator {
	return SoftLayerCreator{
		softLayerClient: client,
		logger:          logger,
	}
}

func (c SoftLayerCreator) Create(size int, virtualGuestId int) (Disk, error) {
	c.logger.Debug(softLayerCreatorLogTag, "Creating disk of size '%d'", size)

	vmService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return SoftLayerDisk{}, bosherr.WrapError(err, "Create SoftLayer Virtual Guest Service error.")
	}

	vm, err := vmService.GetObject(virtualGuestId)
	if err != nil || vm.Id == 0 {
		return SoftLayerDisk{}, bosherr.WrapError(err, fmt.Sprintf("Can not retrieve vitual guest with id: %d.", virtualGuestId))
	}

	storageService, err := c.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return SoftLayerDisk{}, bosherr.WrapError(err, "Create SoftLayer Network Storage Service error.")
	}

	disk, err := storageService.CreateIscsiVolume(size, strconv.Itoa(vm.Datacenter.Id))
	if err != nil {
		return SoftLayerDisk{}, bosherr.WrapError(err, "Create SoftLayer iSCSI disk error.")
	}

	return NewSoftLayerDisk(disk.Id, c.logger), nil
}
