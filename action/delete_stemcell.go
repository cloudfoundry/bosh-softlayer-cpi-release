package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
)

type DeleteStemcellAction struct {
	stemcellFinder bslcstem.StemcellFinder
	logger         boshlog.Logger
}

func NewDeleteStemcell(
	stemcellFinder bslcstem.StemcellFinder,
	logger boshlog.Logger,
) (action DeleteStemcellAction) {
	action.stemcellFinder = stemcellFinder
	action.logger = logger
	return
}

func (a DeleteStemcellAction) Run(stemcellCID StemcellCID) (interface{}, error) {
	return nil, nil
}
