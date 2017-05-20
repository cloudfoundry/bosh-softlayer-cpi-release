package action

import (
	stemcell_service "bosh-softlayer-cpi/softlayer/stemcell_service"
)

type DeleteStemcellAction struct {
	stemcellService stemcell_service.SoftlayerStemcellService
}

func NewDeleteStemcell(
	stemcellService stemcell_service.SoftlayerStemcellService,
) (action DeleteStemcellAction) {
	action.stemcellService = stemcellService
	return
}

func (a DeleteStemcellAction) Run(stemcellCID StemcellCID) (interface{}, error) {
	return nil, nil
}
