package stemcell

import (
	bsl "bosh-softlayer-cpi/softlayer/client"
	"code.cloudfoundry.org/clock"
	"fmt"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
	"time"
)

type SoftLayerStemcellFinder struct {
	client bsl.Client
	logger boshlog.Logger
}

func NewSoftLayerStemcellFinder(client bsl.Client, logger boshlog.Logger) SoftLayerStemcellFinder {
	return SoftLayerStemcellFinder{client: client, logger: logger}
}

func (ssf SoftLayerStemcellFinder) FindById(id int) (Stemcell, error) {
	var vgbdtg datatypes.Virtual_Guest_Block_Device_Template_Group
	var err error
	execStmtRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			vgbdtg, err = ssf.client.GetImage(id, bsl.IMAGE_DEFAULT_MASK)
			if err != nil {
				apiErr := err.(sl.Error)
				if apiErr.Exception == "SoftLayer_Exception_ObjectNotFound" {
					return true, bosherr.WrapErrorf(err, fmt.Sprintf("VirtualGuestBlockDeviceTemplateGroup with id `%d` does not exist", id))
				} else {
					return false, bosherr.WrapErrorf(err, fmt.Sprintf("Getting VirtualGuestBlockDeviceTemplateGroup with id `%d`", id))
				}
			}

			return false, nil
		})
	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(1*time.Minute, 5*time.Second, execStmtRetryable, timeService, boshlog.NewLogger(boshlog.LevelInfo))
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return SoftLayerStemcell{}, bosherr.Error(fmt.Sprintf("Getting VirtualGuestBlockDeviceTemplateGroup with id `%d` timeout after retried", id))
	}

	return NewSoftLayerStemcell(*vgbdtg.Id, *vgbdtg.GlobalIdentifier, ssf.logger), nil
}
