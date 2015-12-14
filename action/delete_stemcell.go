package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

const (
	deleteStemcellLogTag = "DeleteStemcell"
)

type DeleteStemcell struct {
	stemcellFinder bslcstem.Finder
	logger         boshlog.Logger
}

func NewDeleteStemcell(stemcellFinder bslcstem.Finder, logger boshlog.Logger) DeleteStemcell {
	return DeleteStemcell{stemcellFinder: stemcellFinder, logger: logger}
}

func (a DeleteStemcell) Run(stemcellCID StemcellCID) (interface{}, error) {
	_, found, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		a.logger.Info(deleteStemcellLogTag, "Error trying to find stemcell '%s': %s", stemcellCID, err)
	} else if !found {
		a.logger.Info(deleteStemcellLogTag, "Stemcell '%s' not found", stemcellCID)
	}

	return nil, nil
}
