package action

import (
	"bosh-softlayer-cpi/api"
	"bosh-softlayer-cpi/softlayer/stemcell_service"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

const softlayerInfrastructure = "softlayer"
const bluemixInfrastructure = "bluemix"

type CreateStemcellAction struct {
	stemcellService stemcell.Service
}

func NewCreateStemcell(
	stemcellFinder stemcell.Service,
) (action CreateStemcellAction) {
	action.stemcellService = stemcellFinder
	return
}

func (a CreateStemcellAction) Run(imagePath string, cloudProps StemcellCloudProperties) (string, error) {
	var err error
	var stemcell string

	if cloudProps.Infrastructure != softlayerInfrastructure && cloudProps.Infrastructure != bluemixInfrastructure {
		return "", bosherr.Errorf("Create stemcell: Invalid '%s' infrastructure", cloudProps.Infrastructure)
	}

	switch {
	case cloudProps.Id != 0:
		_, err = a.stemcellService.Find(cloudProps.Id)
		if err != nil {
			if _, ok := err.(api.CloudError); ok {
				return "", err
			}
			return "", bosherr.WrapErrorf(err, "Create stemcell from light-stemcell")
		}
		stemcell = StemcellCID(cloudProps.Id).String()
	default:
		stemcellId, err := a.stemcellService.CreateFromTarball(imagePath, cloudProps.DatacenterName, cloudProps.OsCode)
		if err != nil {
			if _, ok := err.(api.CloudError); ok {
				return "", err
			}
			return "", bosherr.WrapErrorf(err, "Create stemcell from raw-stemcell")
		}
		stemcell = StemcellCID(stemcellId).String()
	}

	return stemcell, nil
}
