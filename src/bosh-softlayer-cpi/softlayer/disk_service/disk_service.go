package disk

import "github.com/softlayer/softlayer-go/datatypes"

type Service interface {
	Create(size int, iops int, location string) (int, error)
	Delete(id int) error
	Find(id int) (datatypes.Network_Storage, bool, error)
}
