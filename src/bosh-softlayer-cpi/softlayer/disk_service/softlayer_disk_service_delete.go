package disk

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (d SoftlayerDiskService) Delete(id int) error {
	_, err := d.softlayerClient.CancelBlockVolume(id, "By BOSH !!!", true)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting disk with id '%d'", id)
	}

	return nil
}
