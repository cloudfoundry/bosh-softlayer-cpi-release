package disk

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

const fsDiskLogTag = "FSDisk"

type FSDisk struct {
	id   int
	path string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewFSDisk(
	id int,
	path string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) FSDisk {
	return FSDisk{id: id, path: path, fs: fs, logger: logger}
}

func (s FSDisk) ID() int { return s.id }

func (s FSDisk) Path() string { return s.path }

func (s FSDisk) Delete() error {
	s.logger.Debug(fsDiskLogTag, "Deleting disk '%s'", s.id)

	err := s.fs.RemoveAll(s.path)
	if err != nil {
		return bosherr.WrapError(err, "Deleting disk '%s'", s.path)
	}

	return nil
}
