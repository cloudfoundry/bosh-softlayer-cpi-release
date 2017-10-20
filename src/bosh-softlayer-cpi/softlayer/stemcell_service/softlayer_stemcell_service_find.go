package stemcell

import (
	"code.cloudfoundry.org/clock"
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"

	bosl "bosh-softlayer-cpi/softlayer/client"

	"bosh-softlayer-cpi/api"
	"github.com/softlayer/softlayer-go/datatypes"
	"strconv"
)

func (s SoftlayerStemcellService) Find(id int) (string, error) {
	var (
		vgbdtg *datatypes.Virtual_Guest_Block_Device_Template_Group
		err    error
		found  bool
	)
	execStmtRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			vgbdtg, found, err = s.softlayerClient.GetImage(id, bosl.IMAGE_DEFAULT_MASK)
			if err != nil {
				return true, bosherr.WrapErrorf(err, fmt.Sprintf("Getting VirtualGuestBlockDeviceTemplateGroup with id '%d'", id))
			}

			if !found {
				return false, api.NewStemcellkNotFoundError(strconv.Itoa(id), false)
			}

			return false, nil
		})
	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(1*time.Minute, 5*time.Second, execStmtRetryable, timeService, s.logger.GetbasicLogger())
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return "", err
	}

	return *vgbdtg.GlobalIdentifier, nil
}
