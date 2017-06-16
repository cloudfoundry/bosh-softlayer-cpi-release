package instance

import (
	"bosh-softlayer-cpi/api"
	"github.com/softlayer/softlayer-go/datatypes"
	"strconv"
)

func (vg SoftlayerVirtualGuestService) ReloadOS(id int, stemcellID int, sshKeys []int, vmNamePrefix string, domain string) error {
	return vg.softlayerClient.ReloadInstance(id, stemcellID, sshKeys, vmNamePrefix, domain)
}

func (vg SoftlayerVirtualGuestService) Edit(id int, instance *datatypes.Virtual_Guest) error {
	found, err := vg.softlayerClient.EditInstance(id, instance)
	if err != nil {
		return err
	}

	if !found {
		return api.NewVMNotFoundError(strconv.Itoa(id))
	}

	return nil
}
