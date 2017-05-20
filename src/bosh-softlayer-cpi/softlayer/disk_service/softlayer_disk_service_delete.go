package disk

import (
	"bosh-softlayer-cpi/api"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (d SoftlayerDiskService) Delete(id int) error {
	_, found, err := d.Find(id)
	if err != nil {
		return err
	}
	if !found {
		return api.NewDiskNotFoundError(string(id), false)
	}

	err = d.softlayerClient.CancelBlockVolume(id, "", true)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting disk with id '%d'", id)
	}

	return nil
}
