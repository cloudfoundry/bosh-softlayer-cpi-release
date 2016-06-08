package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
)

const (
	deleteStemcellLogTag = "DeleteStemcell"
)

type DeleteStemcellAction struct {
	stemcellFinder bslcstem.Finder
	logger         boshlog.Logger
}

func NewDeleteStemcell(
	stemcellFinder bslcstem.Finder,
	logger boshlog.Logger,
) (action DeleteStemcellAction) {
	action.stemcellFinder = stemcellFinder
	action.logger = logger
	return
}

func (a DeleteStemcellAction) Run(stemcellCID StemcellCID) (interface{}, error) {
	_, found, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		a.logger.Info(deleteStemcellLogTag, "Error trying to find stemcell '%s': %s", stemcellCID, err)
	} else if !found {
		a.logger.Info(deleteStemcellLogTag, "Stemcell '%s' not found", stemcellCID)
	}

	return nil, nil
}
