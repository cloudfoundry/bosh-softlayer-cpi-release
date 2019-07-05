package action

import (
	bosherr "github.com/bluebosh/bosh-utils/errors"

	"bosh-softlayer-cpi/softlayer/snapshot_service"
)

type DeleteSnapshot struct {
	snapshotService snapshot.Service
}

func NewDeleteSnapshot(
	snapshotService snapshot.Service,
) DeleteSnapshot {
	return DeleteSnapshot{
		snapshotService: snapshotService,
	}
}

func (ds DeleteSnapshot) Run(snapshotCID SnapshotCID) (interface{}, error) {
	if err := ds.snapshotService.Delete(snapshotCID.Int()); err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting snapshot '%d'", snapshotCID.Int())
	}

	return nil, nil
}
