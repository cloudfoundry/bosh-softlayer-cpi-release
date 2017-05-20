package stemcell

import (
	"code.cloudfoundry.org/clock"
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"

	bosl "bosh-softlayer-cpi/softlayer/client"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

func (s SoftlayerStemcellService) Find(id int) (string, bool, error) {
	var vgbdtg datatypes.Virtual_Guest_Block_Device_Template_Group
	var err error
	execStmtRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			vgbdtg, err = s.softlayerClient.GetImage(id, bosl.IMAGE_DEFAULT_MASK)
			if err != nil {
				apiErr := err.(sl.Error)
				if apiErr.Exception == "SoftLayer_Exception_ObjectNotFound" {
					return false, nil
				} else {
					return true, bosherr.WrapErrorf(err, fmt.Sprintf("Getting VirtualGuestBlockDeviceTemplateGroup with id '%d'", id))
				}
			}

			return false, nil
		})
	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(1*time.Minute, 5*time.Second, execStmtRetryable, timeService, s.logger)
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return "", false, bosherr.Error(fmt.Sprintf("Getting VirtualGuestBlockDeviceTemplateGroup with id '%d'", id))
	}

	if vgbdtg.GlobalIdentifier == nil {
		return "", false, nil
	}

	return *vgbdtg.GlobalIdentifier, true, nil
}
