package instance

import bosherr "github.com/bluebosh/bosh-utils/errors"

func (vg SoftlayerVirtualGuestService) Delete(id int, enableVps bool) error {
	_, found, err := vg.softlayerClient.GetInstance(id, "id")
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with id '%d'", id)
	}
	if !found {
		return nil
	}

	if enableVps {
		return vg.softlayerClient.DeleteInstanceFromVPS(id)
	}

	return vg.softlayerClient.CancelInstance(id)
}
