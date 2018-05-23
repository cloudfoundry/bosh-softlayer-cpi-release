package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (vg SoftlayerVirtualGuestService) Reboot(id int) error {
	err := vg.softlayerClient.RebootInstance(id, true, false)
	if err != nil {
		return bosherr.WrapError(err, "Rebooting (soft) SoftLayer VirtualGuest from client")
	}

	return nil
}
