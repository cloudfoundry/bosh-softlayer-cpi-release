package disk

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	slc "github.com/maximilien/softlayer-go/softlayer"
)

const SOFTLAYER_DISK_FINDER_LOG_TAG = "SoftLayerDiskFinder"

type SoftLayerFinder struct {
	softLayerClient slc.Client
	logger          boshlog.Logger
}

func NewSoftLayerDiskFinder(client slc.Client, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{softLayerClient: client, logger: logger}
}

func (f SoftLayerFinder) Find(id int) (Disk, bool, error) {
	f.logger.Debug(SOFTLAYER_DISK_FINDER_LOG_TAG, "Finding disk '%s'", id)

	service, err := f.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Cannot get network storage service.")
	}

	disk, err := service.GetIscsiVolume(id)
	if err != nil {
		return nil, false, bosherr.WrapErrorf(err, "Failed to find iSCSI volume with id: %d", id)
	}

	if disk.Id == 0 {
		return nil, false, nil
	}

	result := NewSoftLayerDisk(id, f.softLayerClient, f.logger)

	return result, true, nil
}
