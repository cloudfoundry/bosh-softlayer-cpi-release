package instance

import "bosh-softlayer-cpi/api"

func (vg SoftlayerVirtualGuestService) Delete(id int, enableVps bool) error {
	_, found, err := vg.Find(id)
	if err != nil {
		return err
	}
	if !found {
		return api.NewVMNotFoundError(string(id))
	}

	if enableVps {
		return vg.softlayerClient.DeleteInstanceFromVPS(id)
	}

	return vg.softlayerClient.CancelInstance(id)
}
