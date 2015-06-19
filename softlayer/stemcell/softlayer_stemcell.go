package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	sl "github.com/maximilien/softlayer-go/softlayer"
)

const softLayerStemcellLogTag = "SoftLayerStemcell"

const VirtualDiskImageKind = "VirtualDiskImage"
const VirtualGuestDeviceTemplateGroupKind = "VirtualGuestDeviceTemplateGroup"
const DefaultKind = VirtualGuestDeviceTemplateGroupKind

type SoftLayerStemcell struct {
	id   int
	uuid string
	kind string

	softLayerClient sl.Client

	logger boshlog.Logger
}

func NewSoftLayerStemcell(id int, uuid string, kind string, softLayerClient sl.Client, logger boshlog.Logger) SoftLayerStemcell {
	return SoftLayerStemcell{
		id:              id,
		uuid:            uuid,
		kind:            kind,
		softLayerClient: softLayerClient,
		logger:          logger,
	}
}

func (s SoftLayerStemcell) ID() int { return s.id }

func (s SoftLayerStemcell) Uuid() string { return s.uuid }

func (s SoftLayerStemcell) Kind() string { return s.kind }

func (s SoftLayerStemcell) Delete() error {
	if s.kind == VirtualGuestDeviceTemplateGroupKind {
		return s.deleteVirtualGuestDiskTemplateGroup(s.id)
	} else if s.kind == VirtualDiskImageKind {
		return nil
	} else {
		return bosherr.WrapError(nil, "Unknown SoftLayer stemcell kind")
	}
}

func (s SoftLayerStemcell) deleteVirtualGuestDiskTemplateGroup(id int) error {
	vgdtgService, err := s.softLayerClient.GetSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service()
	if err != nil {
		return bosherr.WrapError(err, "Getting SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service from SoftLayer client")
	}

	_, err = vgdtgService.DeleteObject(id)
	if err != nil {
		return bosherr.WrapError(err, "Deleting VirtualGuestBlockDeviceTemplateGroup from service")
	}

	//TODO: fix to check that transaction completed since vgdtgService.DeleteObject(id) does not return bool but a transaction
	// if !deleted {
	// 	return bosherr.WrapError(nil, fmt.Sprintf("Could not delete VirtualGuestBlockDeviceTemplateGroup with id `%d`", id))
	// }

	return err
}
