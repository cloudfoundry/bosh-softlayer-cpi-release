package disk

import (
	// bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	// boshsys "github.com/cloudfoundry/bosh-agent/system"
)

const iscsiDiskLogTag = "iSCSIDisk"

type IscsiDisk struct {
	id     int
	logger boshlog.Logger
}

func NewIscsiDisk(id int, logger boshlog.Logger) IscsiDisk {
	return IscsiDisk{id: id, logger: logger}
}

func (s IscsiDisk) ID() int { return s.id }

func (s IscsiDisk) Delete() error {
	s.logger.Debug(iscsiDiskLogTag, "Deleting iSCSI disk '%s'", s.id)

	// if err != nil {
	// 	return bosherr.WrapError(err, "Deleting iSCSI disk '%s'", s.id)
	// }

	return nil
}
