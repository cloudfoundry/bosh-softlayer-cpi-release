package action

import (
	"bosh-softlayer-cpi/api"
	stemcell "bosh-softlayer-cpi/softlayer/stemcell_service"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CreateStemcellAction struct {
	stemcellService stemcell.SoftlayerStemcellService
}

type CreateStemcellCloudProps struct {
	Id             int    `json:"virtual-disk-image-id"`
	Uuid           string `json:"virtual-disk-image-uuid"`
	DatacenterName string `json:"datacenter-name"`
}

func NewCreateStemcell(
	stemcellFinder stemcell.SoftlayerStemcellService,
) (action CreateStemcellAction) {
	action.stemcellService = stemcellFinder
	return
}

func (a CreateStemcellAction) Run(imagePath string, stemcellCloudProps CreateStemcellCloudProps) (string, error) {
	_, found, err := a.stemcellService.Find(stemcellCloudProps.Id)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding stemcell with id '%d'", stemcellCloudProps.Id)
	}

	if !found {
		return "", api.NewStemcellkNotFoundError(string(stemcellCloudProps.Id), false)
	}

	return StemcellCID(stemcellCloudProps.Id).String(), nil
}
