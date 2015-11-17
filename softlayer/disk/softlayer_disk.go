package disk

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	slc "github.com/maximilien/softlayer-go/softlayer"
)

const SOFTLAYER_DISK_LOG_TAG = "SoftLayerDisk"

type SoftLayerDisk struct {
	id              int
	softLayerClient slc.Client
	logger          boshlog.Logger
}

func NewSoftLayerDisk(id int, client slc.Client, logger boshlog.Logger) SoftLayerDisk {
	return SoftLayerDisk{
		id:              id,
		softLayerClient: client,
		logger:          logger,
	}
}

func (s SoftLayerDisk) ID() int { return s.id }

func (s SoftLayerDisk) Delete() error {
	s.logger.Debug(SOFTLAYER_DISK_LOG_TAG, "Deleting disk '%s'", s.id)

	service, err := s.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	err = service.DeleteIscsiVolume(s.id, true)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to delete iSCSI volume with id: %d", s.id)
	}

	return nil
}
