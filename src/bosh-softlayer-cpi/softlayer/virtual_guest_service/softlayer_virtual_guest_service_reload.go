package instance

func (vg SoftlayerVirtualGuestService) ReloadOS(id int, stemcellID int) error {
	return vg.softlayerClient.ReloadInstance(id, stemcellID)
}
