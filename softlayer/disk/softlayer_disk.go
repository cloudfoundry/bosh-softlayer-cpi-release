package disk

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const softLayerDiskLogTag = "SoftLayerDisk"

type SoftLayerDisk struct {
	id     int
	logger boshlog.Logger
}

func NewSoftLayerDisk(id int, logger boshlog.Logger) SoftLayerDisk {
	return SoftLayerDisk{id: id, logger: logger}
}

func (s SoftLayerDisk) ID() int { return s.id }

func (s SoftLayerDisk) Delete() error {
	s.logger.Debug(softLayerDiskLogTag, "Deleting disk '%s'", s.id)
	return nil
}
