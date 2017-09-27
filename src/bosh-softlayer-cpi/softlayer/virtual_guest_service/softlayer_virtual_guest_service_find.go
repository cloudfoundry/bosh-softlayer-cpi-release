package instance

import (
	boslc "bosh-softlayer-cpi/softlayer/client"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"bosh-softlayer-cpi/api"
	"github.com/softlayer/softlayer-go/datatypes"
	"strconv"
)

func (vg SoftlayerVirtualGuestService) Find(id int) (*datatypes.Virtual_Guest, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest '%d' ", id)
	instance, found, err := vg.softlayerClient.GetInstance(id, boslc.INSTANCE_DETAIL_MASK)
	if err != nil {
		return &datatypes.Virtual_Guest{}, bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with id '%d'", id)
	}

	if !found {
		return &datatypes.Virtual_Guest{}, api.NewVMNotFoundError(strconv.Itoa(id))
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
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest by Primary Backend IP Address '%s' ", ip)
	instance, found, err := vg.softlayerClient.GetInstanceByPrimaryIpAddress(ip)
	if err != nil {
		return &datatypes.Virtual_Guest{}, bosherr.WrapErrorf(err, "Failed to find SoftLayer VirtualGuest with ip '%s'", ip)
	}

	if !found {
		return &datatypes.Virtual_Guest{}, api.NewVMNotFoundError(ip)
	}

	return instance, nil
}
