package disk

import (
	// "path/filepath"
	// "strconv"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	// boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type SoftLayerFinder struct {
	logger boshlog.Logger
}

func NewSoftLayerFinder(logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{logger: logger}
}

func (f SoftLayerFinder) Find(id int) (Disk, bool, error) {
	return nil, false, nil
}
