package stemcell

import (
	"os"
	"path/filepath"

	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshcmd "bosh/platform/commands"
	boshsys "bosh/system"
	boshuuid "bosh/uuid"
)

const fsImporterLogTag = "FSImporter"

type FSImporter struct {
	dirPath string

	fs         boshsys.FileSystem
	uuidGen    boshuuid.Generator
	compressor boshcmd.Compressor

	logger boshlog.Logger
}

func NewFSImporter(
	dirPath string,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	compressor boshcmd.Compressor,
	logger boshlog.Logger,
) FSImporter {
	return FSImporter{
		dirPath: dirPath,

		fs:         fs,
		uuidGen:    uuidGen,
		compressor: compressor,

		logger: logger,
	}
}

func (i FSImporter) ImportFromPath(imagePath string) (Stemcell, error) {
	i.logger.Debug(fsImporterLogTag, "Importing stemcell from path '%s'", imagePath)

	id, err := i.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating stemcell id")
	}

	stemcellPath := filepath.Join(i.dirPath, id)

	err = i.fs.MkdirAll(stemcellPath, os.FileMode(0755))
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating stemcell directory '%s'", stemcellPath)
	}

	err = i.compressor.DecompressFileToDir(imagePath, stemcellPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Unpacking stemcell '%s' to '%s'", imagePath, stemcellPath)
	}

	i.logger.Debug(fsImporterLogTag, "Imported stemcell from path '%s'", imagePath)

	return NewFSStemcell(id, stemcellPath, i.fs, i.logger), nil
}
