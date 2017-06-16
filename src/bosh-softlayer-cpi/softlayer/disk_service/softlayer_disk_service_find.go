package disk

import (
	"bosh-softlayer-cpi/api"
	boslc "bosh-softlayer-cpi/softlayer/client"
	"github.com/softlayer/softlayer-go/datatypes"
	"strconv"
)

func (d SoftlayerDiskService) Find(id int) (*datatypes.Network_Storage, error) {
	volume, found, err := d.softlayerClient.GetBlockVolumeDetails(id, boslc.VOLUME_DEFAULT_MASK)
	if err != nil {
		return &datatypes.Network_Storage{}, err
	}

	if !found {
		return &datatypes.Network_Storage{}, api.NewDiskNotFoundError(strconv.Itoa(id), false)
	}

	return volume, nil
}
