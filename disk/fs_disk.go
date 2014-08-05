package disk

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

const fsDiskLogTag = "FSDisk"

type FSDisk struct {
	id   string
	path string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewFSDisk(
	id string,
	path string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) FSDisk {
	return FSDisk{id: id, path: path, fs: fs, logger: logger}
}

func (s FSDisk) ID() string { return s.id }

func (s FSDisk) Path() string { return s.path }

func (s FSDisk) Delete() error {
	s.logger.Debug(fsDiskLogTag, "Deleting disk '%s'", s.id)

	err := s.fs.RemoveAll(s.path)
	if err != nil {
		return bosherr.WrapError(err, "Deleting disk '%s'", s.path)
	}

	return nil
}
