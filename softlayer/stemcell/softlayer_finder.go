package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"
)

type SoftLayerFinder struct {
	client sl.Client
	logger boshlog.Logger
}

func NewSoftLayerFinder(client sl.Client, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{client: client, logger: logger}
}

func (f SoftLayerFinder) Find(id string) (Stemcell, bool, error) {
	accountService, err := f.client.GetSoftLayer_Account_Service()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting SoftLayer AccountService")
	}

	virtualDiskImages, err := accountService.GetVirtualDiskImages()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting virtual disk images")
	}

	for _, vdImage := range virtualDiskImages {
		if vdImage.Uuid == id {
			return NewSoftLayerStemcell(id, f.logger), true, nil
		}
	}

	return nil, false, nil
}
