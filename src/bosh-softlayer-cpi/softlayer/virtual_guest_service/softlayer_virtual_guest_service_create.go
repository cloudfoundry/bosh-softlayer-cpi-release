package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/softlayer/softlayer-go/datatypes"
)

func (vg SoftlayerVirtualGuestService) Create(virtualGuest *datatypes.Virtual_Guest, enableVps bool, stemcellID int, sshKeys []int) (int, error) {
	var err error

	if enableVps {
		virtualGuest, err = vg.softlayerClient.CreateInstanceFromVPS(virtualGuest, stemcellID, sshKeys)
	} else {
		virtualGuest, err = vg.softlayerClient.CreateInstance(virtualGuest)
	}
	if err != nil {
		return 0, bosherr.WrapError(err, "Creating virtualGuest")
	}

	return *virtualGuest.Id, nil
}

func (vg SoftlayerVirtualGuestService) CleanUp(id int) {
	if err := vg.Delete(id, false); err != nil {
		vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Failed cleaning up Softlayer VirtualGuest '%s': %v", id, err)
	}
}
