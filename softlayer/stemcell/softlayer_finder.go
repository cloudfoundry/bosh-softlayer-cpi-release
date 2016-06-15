package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sl_datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"

	"fmt"
	"github.com/pivotal-golang/clock"
	"strings"
)

type SoftLayerFinder struct {
	client sl.Client
	logger boshlog.Logger
}

func NewSoftLayerFinder(client sl.Client, logger boshlog.Logger) SoftLayerFinder {
	return SoftLayerFinder{client: client, logger: logger}
}

func (f SoftLayerFinder) FindById(id int) (Stemcell, error) {
	vgbdtg := sl_datatypes.SoftLayer_Virtual_Guest_Block_Device_Template_Group{}
	vgdtgService, err := f.client.GetSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service()

	execStmtRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			vgbdtg, err = vgdtgService.GetObject(id)
			if err != nil {
				if strings.Contains(err.Error(), "404") {
					return false, bosherr.Error(fmt.Sprintf("Failed to get VirtualGuestBlockDeviceTemplateGroup with id `%d`", id))
				} else {
					return true, bosherr.Error(fmt.Sprintf("The VirtualGuestBlockDeviceTemplateGroup with id `%d` does not exist", id))
				}
			}

			return false, nil
		})
	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL, execStmtRetryable, timeService, boshlog.NewLogger(boshlog.LevelInfo))
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return SoftLayerStemcell{}, bosherr.Error(fmt.Sprintf("Can not find VirtualGuestBlockDeviceTemplateGroup with id `%d`", id))
	}

	return NewSoftLayerStemcell(vgbdtg.Id, vgbdtg.GlobalIdentifier, f.client, f.logger), nil
}
