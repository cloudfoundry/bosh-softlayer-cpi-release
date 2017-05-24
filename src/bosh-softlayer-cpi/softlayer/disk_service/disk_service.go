package disk

import "github.com/softlayer/softlayer-go/datatypes"

//go:generate counterfeiter -o fakes/fake_Disk_Service.go . Service
type Service interface {
	Create(size int, iops int, location string) (int, error)
	Delete(id int) error
	Find(id int) (datatypes.Network_Storage, bool, error)
}
