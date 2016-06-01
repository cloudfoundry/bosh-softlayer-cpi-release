package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"fmt"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

type SoftLayerFinder struct {
	client sl.Client
	logger boshlog.Logger
}

func NewSoftLayerFinder(client sl.Client, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{client: client, logger: logger}
}

func (f SoftLayerFinder) FindById(id int) (Stemcell, bool, error) {
	accountService, err := f.client.GetSoftLayer_Account_Service()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting SoftLayer AccountService")
	}

	stemcell, found, err := f.findByIdInVirtualDiskImages(id, accountService)
	if err != nil {
		return stemcell, found, err
	}

	if found {
		return stemcell, found, nil
	} else {
		stemcell, found, err = f.findByIdInVirtualGuestDeviceTemplateGroups(id, accountService)
		if err != nil {
			return stemcell, found, err
		}
	}

	return stemcell, found, nil
}

func (f SoftLayerFinder) Find(uuid string) (Stemcell, bool, error) {
	accountService, err := f.client.GetSoftLayer_Account_Service()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting SoftLayer AccountService")
	}

	stemcell, found, err := f.findInVirtualDiskImages(uuid, accountService)
	if err != nil {
		return stemcell, found, err
	}

	if found {
		return stemcell, found, nil
	} else {
		stemcell, found, err = f.findInVirtualGuestDeviceTemplateGroups(uuid, accountService)
		if err != nil {
			return stemcell, found, err
		}
	}

	return stemcell, found, nil
}

func (f SoftLayerFinder) findInVirtualDiskImages(uuid string, accountService sl.SoftLayer_Account_Service) (Stemcell, bool, error) {
	filters := fmt.Sprintf(`{"virtualDiskImages":{"uuid":{"operation":"%s"}}}`, uuid)
	virtualDiskImages, err := accountService.GetVirtualDiskImagesWithFilter(filters)
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting virtual disk images")
	}

	for _, vdImage := range virtualDiskImages {
		if vdImage.Uuid == uuid {
			return NewSoftLayerStemcell(vdImage.Id, vdImage.Uuid, VirtualDiskImageKind, f.client, f.logger), true, nil
		}
	}

	return nil, false, nil
}

func (f SoftLayerFinder) findByIdInVirtualDiskImages(id int, accountService sl.SoftLayer_Account_Service) (Stemcell, bool, error) {
	filters := fmt.Sprintf(`{"virtualDiskImages":{"id":{"operation":"%d"}}}`, id)
	virtualDiskImages, err := accountService.GetVirtualDiskImagesWithFilter(filters)
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting virtual disk images")
	}

	for _, vdImage := range virtualDiskImages {
		if vdImage.Id == id {
			return NewSoftLayerStemcell(vdImage.Id, vdImage.Uuid, VirtualDiskImageKind, f.client, f.logger), true, nil
		}
	}

	return nil, false, nil
}

func (f SoftLayerFinder) findInVirtualGuestDeviceTemplateGroups(uuid string, accountService sl.SoftLayer_Account_Service) (Stemcell, bool, error) {
	filters := fmt.Sprintf(`{"blockDeviceTemplateGroups":{"globalIdentifier":{"operation":"%s"}}}`, uuid)
	vgdtgGroups, err := accountService.GetBlockDeviceTemplateGroupsWithFilter(filters)
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting virtual guest device template groups")
	}

	for _, vgdtgGroup := range vgdtgGroups {
		if vgdtgGroup.GlobalIdentifier == uuid {
			return NewSoftLayerStemcell(vgdtgGroup.Id, vgdtgGroup.GlobalIdentifier, VirtualGuestDeviceTemplateGroupKind, f.client, f.logger), true, nil
		}
	}

	return nil, false, nil
}

func (f SoftLayerFinder) findByIdInVirtualGuestDeviceTemplateGroups(id int, accountService sl.SoftLayer_Account_Service) (Stemcell, bool, error) {
	filters := fmt.Sprintf(`{"blockDeviceTemplateGroups":{"id":{"operation":"%d"}}}`, id)
	vgdtgGroups, err := accountService.GetBlockDeviceTemplateGroupsWithFilter(filters)
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Getting virtual guest device template groups")
	}

	for _, vgdtgGroup := range vgdtgGroups {
		if vgdtgGroup.Id == id {
			return NewSoftLayerStemcell(vgdtgGroup.Id, vgdtgGroup.GlobalIdentifier, VirtualGuestDeviceTemplateGroupKind, f.client, f.logger), true, nil
		}
	}

	return nil, false, nil
}
