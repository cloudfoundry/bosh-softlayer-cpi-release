package action

import (
	"bosh-softlayer-cpi/softlayer/disk_service"
)

type HasDisk struct {
	diskService disk.Service
}

func NewHasDisk(
	diskService disk.Service,
) HasDisk {
	return HasDisk{
		diskService: diskService,
	}
}

func (hd HasDisk) Run(diskCID DiskCID) (bool, error) {
	_, err := hd.diskService.Find(diskCID.Int())
	if err != nil {
		return false, err
	}

	return true, nil
}
