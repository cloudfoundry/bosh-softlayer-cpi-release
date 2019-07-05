package action

import (
	"fmt"

	bosherr "github.com/bluebosh/bosh-utils/errors"

	"bosh-softlayer-cpi/softlayer/disk_service"
	"bosh-softlayer-cpi/softlayer/snapshot_service"
)

type SnapshotDisk struct {
	snapshotService snapshot.Service
	diskService     disk.Service
}

func NewSnapshotDisk(
	snapshotService snapshot.Service,
	diskService disk.Service,
) SnapshotDisk {
	return SnapshotDisk{
		snapshotService: snapshotService,
		diskService:     diskService,
	}
}

func (sd SnapshotDisk) Run(diskCID DiskCID, metadata SnapshotMetadata) (string, error) {
	var note string
	// Find the disk
	_, err := sd.diskService.Find(diskCID.Int())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to find disk '%s'", diskCID)
	}

	// Create the disk snapshot
	if metadata.Deployment != "" && metadata.Job != "" && metadata.Index != "" {
		note = fmt.Sprintf("%s_%s_%s", metadata.Deployment, metadata.Job, metadata.Index)
	}

	snapshot, err := sd.snapshotService.Create(diskCID.Int(), note)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating snapshot for disk '%s'", diskCID)
	}

	return DiskCID(snapshot).String(), nil
}
