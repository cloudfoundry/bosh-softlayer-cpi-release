package instance

import "bosh-softlayer-cpi/api"

func (vg SoftlayerVirtualGuestService) Delete(id int) error {
	_, found, err := vg.Find(id)
	if err != nil {
		return err
	}
	if !found {
		return api.NewVMNotFoundError(string(id))
	}

	return vg.softlayerClient.CancelInstance(id)
}
