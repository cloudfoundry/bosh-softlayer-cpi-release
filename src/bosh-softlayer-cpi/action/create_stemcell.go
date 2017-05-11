package action

import (
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CreateStemcellAction struct {
	stemcellFinder bslcstem.StemcellFinder
}

type CreateStemcellCloudProps struct {
	Id             int    `json:"virtual-disk-image-id"`
	Uuid           string `json:"virtual-disk-image-uuid"`
	DatacenterName string `json:"datacenter-name"`
}

func NewCreateStemcell(
	stemcellFinder bslcstem.StemcellFinder,
) (action CreateStemcellAction) {
	action.stemcellFinder = stemcellFinder
	return
}

func (a CreateStemcellAction) Run(imagePath string, stemcellCloudProps CreateStemcellCloudProps) (string, error) {
	stemcell, err := a.stemcellFinder.FindById(stemcellCloudProps.Id)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding stemcell with id '%d'", stemcellCloudProps.Id)
	}

	return StemcellCID(stemcell.ID()).String(), nil
}
