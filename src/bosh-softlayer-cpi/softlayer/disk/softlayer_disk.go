package disk

import (
	bsl "bosh-softlayer-cpi/softlayer/client"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SoftLayerDisk struct {
	id              int
	softLayerClient bsl.Client
	logger          boshlog.Logger
}

func NewSoftLayerDisk(id int, client bsl.Client, logger boshlog.Logger) SoftLayerDisk {
	return SoftLayerDisk{
		id:              id,
		softLayerClient: client,
		logger:          logger,
	}
}

func (sd SoftLayerDisk) ID() int { return sd.id }

func (sd SoftLayerDisk) Delete() error {
	err := sd.softLayerClient.CancelBlockVolume(sd.id, "", true)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting disk with id `%d`", sd.id)
	}

	return nil
}
