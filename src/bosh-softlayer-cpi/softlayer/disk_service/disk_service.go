package disk

import "github.com/softlayer/softlayer-go/datatypes"

//go:generate counterfeiter -o fakes/fake_Disk_Service.go . Service
type Service interface {
	Create(size int, iops int, location string, snapshotSpace int) (int, error)
	Delete(id int) error
	SetMetadata(id int, diskMetadata Metadata) error
	Find(id int) (*datatypes.Network_Storage, error)
}

type Metadata map[string]interface{}
