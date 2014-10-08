package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const softLayerImporterLogTag = "SoftLayerImporter"

type SoftLayerImporter struct {
	dirPath string

	logger boshlog.Logger
}

func NewSoftLayerImporter(logger boshlog.Logger) SoftLayerImporter {
	return SoftLayerImporter{
		logger: logger,
	}
}

func (i SoftLayerImporter) ImportFromPath(imagePath string) (Stemcell, error) {
	i.logger.Debug(softLayerImporterLogTag, "Importing stemcell from path '%s'", imagePath)

	stemcellId := "stemcell-id" //TODO: need to find this from CloudProperties

	return NewSoftLayerStemcell(stemcellId, i.logger), nil
}
