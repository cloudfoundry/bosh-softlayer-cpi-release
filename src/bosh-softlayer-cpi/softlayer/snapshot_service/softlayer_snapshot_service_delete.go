package snapshot

func (s SoftlayerSnapshotService) Delete(id int) error {
	return s.softlayerClient.DeleteSnapshot(id)
}
