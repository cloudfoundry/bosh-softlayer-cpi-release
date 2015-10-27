package baremetal

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

type BaremetalFinder struct {
	client sl.Client
	logger boshlog.Logger
}

func NewBaremetalFinder(client sl.Client, logger boshlog.Logger) BaremetalFinder {
	return BaremetalFinder{client: client, logger: logger}
}

func (f BaremetalFinder) Find(id string) (datatypes.SoftLayer_Hardware, error) {
	service, err := f.client.GetSoftLayer_Hardware_Service()
	if err != nil {
		return datatypes.SoftLayer_Hardware{}, bosherr.WrapError(err, "Get hardware service error")
	}

	baremetal, err := service.GetObject(id)
	if err != nil {
		return datatypes.SoftLayer_Hardware{}, bosherr.WrapError(err, "Get baremetal error")
	}

	if baremetal.GlobalIdentifier == "" {
		return datatypes.SoftLayer_Hardware{}, bosherr.Errorf("cannot find the baremetal server with id: %s.", id)
	}

	return baremetal, nil
}
