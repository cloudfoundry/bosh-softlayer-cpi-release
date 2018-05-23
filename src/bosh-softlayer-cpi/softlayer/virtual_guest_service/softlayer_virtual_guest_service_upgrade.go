package instance

func (vg SoftlayerVirtualGuestService) UpgradeInstance(id int, cpu int, memory int, network int, privateCPU bool) error {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Upgrade instance settings")
	return vg.softlayerClient.UpgradeInstanceConfig(id, cpu, memory, network, privateCPU)
}
