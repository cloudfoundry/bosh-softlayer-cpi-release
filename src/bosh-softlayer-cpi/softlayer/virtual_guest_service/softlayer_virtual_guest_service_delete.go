package instance

func (vg SoftlayerVirtualGuestService) Delete(id int, enableVps bool) error {
	if enableVps {
		return vg.softlayerClient.DeleteInstanceFromVPS(id)
	}

	return vg.softlayerClient.CancelInstance(id)
}
