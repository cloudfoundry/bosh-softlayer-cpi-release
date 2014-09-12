package disk

import (
	"path/filepath"
	"strconv"

	boshlog "bosh/logger"
	boshsys "bosh/system"
)

type FSFinder struct {
	dirPath string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewFSFinder(dirPath string, fs boshsys.FileSystem, logger boshlog.Logger) FSFinder {
	return FSFinder{dirPath: dirPath, fs: fs, logger: logger}
}

func (f FSFinder) Find(id int) (Disk, bool, error) {
	dirPath := filepath.Join(f.dirPath, strconv.Itoa(id))

	if f.fs.FileExists(dirPath) {
		return NewFSDisk(id, dirPath, f.fs, f.logger), true, nil
	}

	return nil, false, nil
}
