package snapshot

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (s SoftlayerSnapshotService) Create(diskID int, note string) (int, error) {
	if note == "" {
		note = softlayerSnapshotDescription
	}

	s.logger.Debug(softlayerSnapshotServiceLogTag, "Creating Softlayer Snapshot with note: %s", note)
	snapshot, err := s.softlayerClient.CreateSnapshot(diskID, note)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Failed to create Softlayer Snapshot")
	}

	return *snapshot.Id, nil
}
