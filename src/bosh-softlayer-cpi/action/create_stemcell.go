package action

import (
	"bosh-softlayer-cpi/api"
	stemcell "bosh-softlayer-cpi/softlayer/stemcell_service"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CreateStemcellAction struct {
	stemcellService stemcell.Service
}

type CreateStemcellCloudProps struct {
	Id             int    `json:"virtual-disk-image-id"`
	Uuid           string `json:"virtual-disk-image-uuid"`
	DatacenterName string `json:"datacenter-name"`
}

func NewCreateStemcell(
	stemcellFinder stemcell.Service,
) (action CreateStemcellAction) {
	action.stemcellService = stemcellFinder
	return
}

func (a CreateStemcellAction) Run(imagePath string, stemcellCloudProps CreateStemcellCloudProps) (string, error) {
	_, err := a.stemcellService.Find(stemcellCloudProps.Id)
	if err != nil {
		if _, ok := err.(api.CloudError); ok {
			return "", err
		}
		return "", bosherr.WrapErrorf(err, "Creating stemcell")
	}

	return StemcellCID(stemcellCloudProps.Id).String(), nil
}
