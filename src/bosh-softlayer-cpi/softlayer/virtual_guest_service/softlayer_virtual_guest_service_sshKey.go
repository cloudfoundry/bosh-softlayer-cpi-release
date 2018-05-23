package instance

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"fmt"
	"github.com/softlayer/softlayer-go/sl"
)

func (vg SoftlayerVirtualGuestService) CreateSshKey(label string, key string, fingerPrint string) (int, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Creating Ssh Public Key with label prefix '%s' ", label)
	uuidStr, err := vg.uuidGen.Generate()
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Generating random SoftLayer ssh public key label")
	}
	sshKey, err := vg.softlayerClient.CreateSshKey(sl.String(fmt.Sprintf("%s_%s", label, uuidStr)), sl.String(key), sl.String(fingerPrint))
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Creating Ssh Public Key with label '%s', key '%s'", label, key)
	}

	return *sshKey.Id, nil
}

func (vg SoftlayerVirtualGuestService) DeleteSshKey(id int) error {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Deleting Ssh Key with id '%d' ", id)
	resp, err := vg.softlayerClient.DeleteSshKey(id)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting Ssh Public Key with id '%d'", id)
	}

	if !resp {
		return bosherr.Errorf("Failed to delete Ssh Public Key with id '%d'", id)
	}

	return nil
}
