package action

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

type CreateStemcell struct {
	stemcellFinder bslcstem.Finder
}

type CreateStemcellCloudProps struct {
	Uuid string `json:"uuid"`
}

func NewCreateStemcell(stemcellFinder bslcstem.Finder) CreateStemcell {
	return CreateStemcell{stemcellFinder: stemcellFinder}
}

func (a CreateStemcell) Run(stemcellCloudProps CreateStemcellCloudProps) (StemcellCID, error) {
	stemcell, found, err := a.stemcellFinder.Find(stemcellCloudProps.Uuid)
	if err != nil {
		return "", bosherr.WrapError(err, "Finding stemcell with UUID '%s'", stemcellCloudProps.Uuid)
	}

	if !found {
		return "", bosherr.WrapError(err, "Did not find stemcell with UUID '%s'", stemcellCloudProps.Uuid)
	}

	return StemcellCID(stemcell.ID()), nil
}
