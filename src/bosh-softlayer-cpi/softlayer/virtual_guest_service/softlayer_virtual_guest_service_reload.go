package instance

import "github.com/softlayer/softlayer-go/datatypes"

func (vg SoftlayerVirtualGuestService) ReloadOS(id int, stemcellID int, sshKeys []int, vmNamePrefix string, domain string) (string, error) {
	return vg.softlayerClient.ReloadInstance(id, stemcellID, sshKeys, vmNamePrefix, domain)
}

func (vg SoftlayerVirtualGuestService) Edit(id int, instance datatypes.Virtual_Guest) (bool, error) {
	return vg.softlayerClient.EditInstance(id, &instance)
}
