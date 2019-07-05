package disk

import (
	bosherr "github.com/bluebosh/bosh-utils/errors"
	"strings"
)

func (d SoftlayerDiskService) Delete(id int) error {
	_, err := d.softlayerClient.CancelBlockVolume(id, "By BOSH !!!", true)
	if err != nil {
		if strings.Contains(err.Error(), "No billing item is found to cancel") {
			return nil
		}
		return bosherr.WrapErrorf(err, "Deleting disk with id '%d'", id)
	}

	return nil
}
