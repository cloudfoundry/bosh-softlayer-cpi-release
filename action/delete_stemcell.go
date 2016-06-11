package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	"time"
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
	bslcommon.TIMEOUT = 30 * time.Second
	bslcommon.POLLING_INTERVAL = 5 * time.Second

	_, err := a.stemcellFinder.FindById(int(stemcellCID))
	if err != nil {
		a.logger.Info(deleteStemcellLogTag, "Stemcell '%s' not found: %s", stemcellCID, err)
	}

	return nil, nil
}
