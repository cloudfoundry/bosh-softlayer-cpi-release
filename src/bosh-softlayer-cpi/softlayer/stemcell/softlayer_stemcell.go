package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"

	sl "github.com/maximilien/softlayer-go/softlayer"

	"fmt"
	"time"
)

type SoftLayerStemcell struct {
	id   int
	uuid string

	softLayerFinder SoftLayerStemcellFinder
}

func NewSoftLayerStemcell(id int, uuid string, softLayerClient sl.Client, logger boshlog.Logger) SoftLayerStemcell {
	slh.TIMEOUT = 60 * time.Minute
	slh.POLLING_INTERVAL = 10 * time.Second

	softLayerFinder := SoftLayerStemcellFinder{
		client: softLayerClient,
		logger: logger,
	}

	return SoftLayerStemcell{
		id:              id,
		uuid:            uuid,
		softLayerFinder: softLayerFinder,
	}
}

func (s SoftLayerStemcell) ID() int { return s.id }

func (s SoftLayerStemcell) Uuid() string { return s.uuid }

func (s SoftLayerStemcell) Delete() error {
	vgdtgService, err := s.softLayerFinder.client.GetSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service()
	if err != nil {
		return bosherr.WrapError(err, "Getting SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service from SoftLayer client")
	}

	_, err = vgdtgService.DeleteObject(s.id)
	if err != nil {
		return bosherr.WrapError(err, "Deleting VirtualGuestBlockDeviceTemplateGroup from service")
	}

	err = slh.WaitForVirtualGuestToHaveNoRunningTransactions(s.softLayerFinder.client, s.id)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d` to have no pending transactions", s.id))
	}

	_, err = s.softLayerFinder.FindById(s.id)
	if err == nil {
		return bosherr.WrapError(nil, fmt.Sprintf("Could not delete VirtualGuestBlockDeviceTemplateGroup with id `%d`", s.id))
	}

	return nil
}
