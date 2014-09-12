package stemcell

import (
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type FSFinder struct {
	dirPath string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewFSFinder(dirPath string, fs boshsys.FileSystem, logger boshlog.Logger) FSFinder {
	return FSFinder{dirPath: dirPath, fs: fs, logger: logger}
}

func (f FSFinder) Find(id string) (Stemcell, bool, error) {
	dirPath := filepath.Join(f.dirPath, id)

	if f.fs.FileExists(dirPath) {
		return NewFSStemcell(id, dirPath, f.fs, f.logger), true, nil
	}

	return nil, false, nil
}
