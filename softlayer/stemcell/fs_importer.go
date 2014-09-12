package stemcell

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
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
	compressor boshcmd.Compressor,
	logger boshlog.Logger,
) FSImporter {
	return FSImporter{
		dirPath: dirPath,

		fs:         fs,
		compressor: compressor,

		logger: logger,
	}
}

func (i FSImporter) ImportFromPath(imagePath string) (Stemcell, error) {
	i.logger.Debug(fsImporterLogTag, "Importing stemcell from path '%s'", imagePath)

	stemcellId := "stemcell-id" //TODO: need to find this from CloudProperties

	stemcellPath := filepath.Join(i.dirPath, stemcellId) 

	err := i.fs.MkdirAll(stemcellPath, os.FileMode(0755))
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating stemcell directory '%s'", stemcellPath)
	}

	err = i.compressor.DecompressFileToDir(imagePath, stemcellPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Unpacking stemcell '%s' to '%s'", imagePath, stemcellPath)
	}

	i.logger.Debug(fsImporterLogTag, "Imported stemcell from path '%s'", imagePath)

	return NewFSStemcell(stemcellId, stemcellPath, i.fs, i.logger), nil
}
