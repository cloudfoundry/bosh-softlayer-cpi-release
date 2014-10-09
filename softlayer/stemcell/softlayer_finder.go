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

	stemcell, found, err := f.findInVirtualDiskImages(id, accountService)
	if err != nil {
		return stemcell, found, err
	}

	if found {
		return stemcell, found, nil
	} else {
		stemcell, found, err = f.findInVirtualGuestDeviceTemplateGroups(id, accountService)
		if err != nil {
			return stemcell, found, err
		}
	}

	return stemcell, found, nil
}

func (f SoftLayerFinder) findInVirtualDiskImages(id string, accountService sl.SoftLayer_Account_Service) (Stemcell, bool, error) {
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

func (f SoftLayerFinder) findInVirtualGuestDeviceTemplateGroups(id string, accountService sl.SoftLayer_Account_Service) (Stemcell, bool, error) {
	vgdtgGroups, err := accountService.GetBlockDeviceTemplateGroups()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting virtual guest device template groups")
	}

	for _, vgdtgGroup := range vgdtgGroups {
		if vgdtgGroup.GlobalIdentifier == id {
			return NewSoftLayerStemcell(id, f.logger), true, nil
		}
	}

	return nil, false, nil
}
