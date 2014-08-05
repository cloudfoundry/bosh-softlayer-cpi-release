package stemcell

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshsys "bosh/system"
)

const fsStemcellLogTag = "FSStemcell"

type FSStemcell struct {
	id      string
	dirPath string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewFSStemcell(id string, dirPath string, fs boshsys.FileSystem, logger boshlog.Logger) FSStemcell {
	return FSStemcell{id: id, dirPath: dirPath, fs: fs, logger: logger}
}

func (s FSStemcell) ID() string { return s.id }

func (s FSStemcell) DirPath() string { return s.dirPath }

func (s FSStemcell) Delete() error {
	s.logger.Debug(fsStemcellLogTag, "Deleting stemcell '%s'", s.id)

	err := s.fs.RemoveAll(s.dirPath)
	if err != nil {
		return bosherr.WrapError(err, "Deleting stemcell directory '%s'", s.dirPath)
	}

	return nil
}
