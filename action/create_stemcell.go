package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

type CreateStemcell struct {
	stemcellFinder bslcstem.Finder
}

type CreateStemcellCloudProps struct {
	Id             int    `json:"virtual-disk-image-id"`
	Uuid           string `json:"virtual-disk-image-uuid"`
	DatacenterName string `json:"datacenter-name"`
}

func NewCreateStemcell(stemcellFinder bslcstem.Finder) CreateStemcell {
	return CreateStemcell{stemcellFinder: stemcellFinder}
}

func (a CreateStemcell) Run(imagePath string, stemcellCloudProps CreateStemcellCloudProps) (string, error) {
	stemcell, found, err := a.stemcellFinder.FindById(stemcellCloudProps.Id)
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Finding stemcell with ID '%d'", stemcellCloudProps.Id)
	}

	if !found {
		return "0", bosherr.Errorf("Did not find stemcell with ID '%d'", stemcellCloudProps.Id)
	}

	return StemcellCID(stemcell.ID()).String(), nil
}
