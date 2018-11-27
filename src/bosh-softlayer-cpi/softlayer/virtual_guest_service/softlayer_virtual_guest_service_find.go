package instance

import (
	"strconv"
	"time"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	"github.com/softlayer/softlayer-go/datatypes"

	"bosh-softlayer-cpi/api"
)

func (vg SoftlayerVirtualGuestService) Find(id int) (*datatypes.Virtual_Guest, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest '%d' ", id)
	var (
		instance *datatypes.Virtual_Guest
		err      error
		found    bool
	)
	//@TODO: Need to encapsulate with stemcell find method
	execStmtRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			instance, found, err = vg.softlayerClient.GetInstance(id, "id, datacenter[name], primaryBackendIpAddress, fullyQualifiedDomainName")
			if err != nil {
				return true, bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with id '%d'", id)
			}

			if !found {
				return false, api.NewVMNotFoundError(strconv.Itoa(id))
			}

			return false, nil
		})
	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(1*time.Minute, 5*time.Second, execStmtRetryable, timeService, vg.logger.GetBoshLogger())
	err = vg.logger.ChangeRetryStrategyLogTag(&timeoutRetryStrategy)
	if err != nil {
		return &datatypes.Virtual_Guest{}, err
	}

	err = timeoutRetryStrategy.Try()
	if err != nil {
		return &datatypes.Virtual_Guest{}, err
	}

	return instance, nil
}

func (vg SoftlayerVirtualGuestService) FindByPrimaryBackendIp(ip string) (*datatypes.Virtual_Guest, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest by Primary Backend IP Address '%s' ", ip)
	instance, found, err := vg.softlayerClient.GetInstanceByPrimaryBackendIpAddress(ip)
	if err != nil {
		return &datatypes.Virtual_Guest{}, bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with ip '%s'", ip)
	}

	if !found {
		return &datatypes.Virtual_Guest{}, api.NewVMNotFoundError(ip)
	}

	return instance, nil
}

func (vg SoftlayerVirtualGuestService) FindByPrimaryIp(ip string) (*datatypes.Virtual_Guest, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest by Primary IP Address '%s' ", ip)
	instance, found, err := vg.softlayerClient.GetInstanceByPrimaryIpAddress(ip)
	if err != nil {
		return &datatypes.Virtual_Guest{}, bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with ip '%s'", ip)
	}

	if !found {
		return &datatypes.Virtual_Guest{}, api.NewVMNotFoundError(ip)
	}

	return instance, nil
}
