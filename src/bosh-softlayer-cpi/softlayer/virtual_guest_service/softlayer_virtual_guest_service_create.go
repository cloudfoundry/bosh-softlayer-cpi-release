package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boslc "bosh-softlayer-cpi/softlayer/client"
	"github.com/softlayer/softlayer-go/datatypes"
)

func (vg SoftlayerVirtualGuestService) Create(virtualGuest datatypes.Virtual_Guest, networks Networks, registryEndpoint string) (int, error) {
	virtualGuest, err := vg.softlayerClient.CreateInstance(&virtualGuest)
	if err != nil {
		return 0, bosherr.WrapError(err, "Creating virtualGuest")
	}

	virtualGuest, err = vg.softlayerClient.GetInstance(*virtualGuest.Id, boslc.INSTANCE_DETAIL_MASK)
	if err != nil {
		return 0, bosherr.WrapError(err, "Getting virtualGuest")
	}

	return *virtualGuest.Id, nil
}

func (vg SoftlayerVirtualGuestService) CleanUp(id int) {
	if err := vg.Delete(id); err != nil {
		vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Failed cleaning up Softlayer VirtualGuest '%s': %v", id, err)
	}

}
