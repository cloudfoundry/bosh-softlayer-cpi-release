package action

import (
	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"time"
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
	bslcommon.TIMEOUT = 30 * time.Second
	bslcommon.POLLING_INTERVAL = 5 * time.Second

	stemcell, err := a.stemcellFinder.FindById(stemcellCloudProps.Id)
	if err != nil {
		return "0", bosherr.WrapErrorf(err, "Finding stemcell with ID '%d'", stemcellCloudProps.Id)
	}

	return StemcellCID(stemcell.ID()).String(), nil
}
