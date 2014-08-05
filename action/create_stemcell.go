package action

import (
	bosherr "bosh/errors"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

type CreateStemcell struct {
	stemcellImporter bslcstem.Importer
}

type CreateStemcellCloudProps struct{}

func NewCreateStemcell(stemcellImporter bslcstem.Importer) CreateStemcell {
	return CreateStemcell{stemcellImporter: stemcellImporter}
}

func (a CreateStemcell) Run(imagePath string, _ CreateStemcellCloudProps) (StemcellCID, error) {
	stemcell, err := a.stemcellImporter.ImportFromPath(imagePath)
	if err != nil {
		return "", bosherr.WrapError(err, "Importing stemcell from '%s'", imagePath)
	}

	return StemcellCID(stemcell.ID()), nil
}
