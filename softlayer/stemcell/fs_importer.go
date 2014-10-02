package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const fsImporterLogTag = "FSImporter"

type FSImporter struct {
	dirPath string

	logger boshlog.Logger
}

func NewFSImporter(
	logger boshlog.Logger,
) FSImporter {
	return FSImporter{
		logger: logger,
	}
}

func (i FSImporter) ImportFromPath(imagePath string) (Stemcell, error) {
	i.logger.Debug(fsImporterLogTag, "Importing stemcell from path '%s'", imagePath)

	stemcellId := "stemcell-id" //TODO: need to find this from CloudProperties

	return NewFSStemcell(stemcellId, i.logger), nil
}
