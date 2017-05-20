package disk

import (
	boslc "bosh-softlayer-cpi/softlayer/client"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

func (d SoftlayerDiskService) Find(id int) (datatypes.Network_Storage, bool, error) {
	volume, err := d.softlayerClient.GetBlockVolumeDetails(id, boslc.VOLUME_DEFAULT_MASK)
	if err != nil {
		apiErr := err.(sl.Error)
		if apiErr.Exception == "SoftLayer_Exception_ObjectNotFound" {
			return datatypes.Network_Storage{}, false, nil
		}

		return datatypes.Network_Storage{}, false, err
	}

	return volume, true, nil
}
