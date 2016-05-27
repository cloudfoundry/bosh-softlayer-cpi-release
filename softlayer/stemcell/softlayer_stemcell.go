package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	"fmt"
	"time"
)

const (
	SOFTLAYER_STEMCELL_LOG_TAG = "SoftLayerStemcell"

	VirtualDiskImageKind                = "VirtualDiskImage"
	VirtualGuestDeviceTemplateGroupKind = "VirtualGuestDeviceTemplateGroup"
	DefaultKind                         = VirtualGuestDeviceTemplateGroupKind
)

type SoftLayerStemcell struct {
	id   int
	uuid string
	kind string

	softLayerClient sl.Client

	logger boshlog.Logger
}

func NewSoftLayerStemcell(id int, uuid string, kind string, softLayerClient sl.Client, logger boshlog.Logger) SoftLayerStemcell {
	bslcommon.TIMEOUT = 60 * time.Minute
	bslcommon.POLLING_INTERVAL = 10 * time.Second

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

	err = slh.WaitForVirtualGuestToHaveNoRunningTransactions(s.softLayerClient, id)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d` to have no pending transactions", id))
	}

	_, err = vgdtgService.GetObject(id)
	if err == nil {
		return bosherr.WrapError(nil, fmt.Sprintf("Could not delete VirtualGuestBlockDeviceTemplateGroup with id `%d`", id))
	}

	return nil
}
