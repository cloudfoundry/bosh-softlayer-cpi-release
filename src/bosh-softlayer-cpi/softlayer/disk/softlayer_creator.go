package disk

import (
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const SOFTLAYER_DISK_CREATOR_LOG_TAG = "SoftLayerDiskCreator"

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

func (c SoftLayerCreator) Create(size int, cloudProps DiskCloudProperties, datacenter_id int) (Disk, error) {
	c.logger.Debug(SOFTLAYER_DISK_CREATOR_LOG_TAG, "Creating disk of size '%d'", size)

	storageService, err := c.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return SoftLayerDisk{}, bosherr.WrapError(err, "Create SoftLayer Network Storage Service error.")
	}

	disk, err := storageService.CreateNetworkStorage(c.getSoftLayerDiskSize(size), cloudProps.Iops, strconv.Itoa(datacenter_id), cloudProps.UseHourlyPricing)
	if err != nil {
		return SoftLayerDisk{}, bosherr.WrapError(err, "Create SoftLayer iSCSI disk error.")
	}

	return NewSoftLayerDisk(disk.Id, c.softLayerClient, c.logger), nil
}

func (c SoftLayerCreator) getSoftLayerDiskSize(size int) int {
	// Sizes and IOPS ranges: http://knowledgelayer.softlayer.com/learning/performance-storage-concepts
	sizeArray := []int{20, 40, 80, 100, 250, 500, 1000, 2000, 4000, 8000, 12000}

	for _, value := range sizeArray {
		if ret := size / 1024; ret <= value {
			return value
		}
	}
	return 12000
}
