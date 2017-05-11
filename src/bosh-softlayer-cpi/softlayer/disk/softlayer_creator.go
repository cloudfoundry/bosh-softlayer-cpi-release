package disk

import (
	bsl "bosh-softlayer-cpi/softlayer/client"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const SOFTLAYER_DISK_CREATOR_LOG_TAG = "SoftLayerDiskCreator"

type SoftLayerCreator struct {
	softLayerClient bsl.Client
	logger          boshlog.Logger
}

func NewSoftLayerDiskCreator(client bsl.Client, logger boshlog.Logger) SoftLayerCreator {
	return SoftLayerCreator{
		softLayerClient: client,
		logger:          logger,
	}
}

func (sc SoftLayerCreator) Create(size int, cloudProps DiskCloudProperties, location string) (Disk, error) {
	sc.logger.Debug(SOFTLAYER_DISK_CREATOR_LOG_TAG, "Creating disk of size '%d'", size)

	volume, err := sc.softLayerClient.CreateVolume(location, sc.getSoftLayerDiskSize(size), cloudProps.Iops)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating volume with size `%d`, iops `%d`, location `%sd`", sc.getSoftLayerDiskSize(size), cloudProps.Iops, location)
	}

	return NewSoftLayerDisk(*volume.Id, sc.softLayerClient, sc.logger), nil
}

func (sc SoftLayerCreator) getSoftLayerDiskSize(size int) int {
	// Sizes and IOPS ranges: http://knowledgelayer.softlayer.com/learning/performance-storage-concepts
	sizeArray := []int{20, 40, 80, 100, 250, 500, 1000, 2000, 4000, 8000, 12000}

	for _, value := range sizeArray {
		if ret := size / 1024; ret <= value {
			return value
		}
	}
	return 12000
}
