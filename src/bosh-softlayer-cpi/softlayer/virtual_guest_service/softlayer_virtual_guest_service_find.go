package instance

import (
	boslc "bosh-softlayer-cpi/softlayer/client"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

func (vg SoftlayerVirtualGuestService) Find(id int) (datatypes.Virtual_Guest, bool, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest '%d' ", id)
	instance, err := vg.softlayerClient.GetInstance(id, boslc.INSTANCE_DETAIL_MASK)
	if err != nil {
		if slErr, ok := err.(sl.Error); ok && slErr.Exception == "SoftLayer_Exception_ObjectNotFound" {
			return datatypes.Virtual_Guest{}, false, nil
		}

		return datatypes.Virtual_Guest{}, false, bosherr.WrapErrorf(err, "Failed to find Softlayer VirtualGuest with id '%d'", id)
	}

	return instance, true, nil
}

func (vg SoftlayerVirtualGuestService) FindByPrimaryBackendIp(ip string) (datatypes.Virtual_Guest, bool, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest by Primary Backend IP Address '%s' ", ip)
	instance, err := vg.softlayerClient.GetInstanceByPrimaryBackendIpAddress(ip)
	if err != nil {
		if slErr, ok := err.(sl.Error); ok && slErr.Exception == "SoftLayer_Exception_ObjectNotFound" {
			return datatypes.Virtual_Guest{}, false, nil
		}

		return datatypes.Virtual_Guest{}, false, bosherr.WrapErrorf(err, "Failed to find Softlayer VirtualGuest with ip '%s'", ip)
	}

	return instance, true, nil
}

func (vg SoftlayerVirtualGuestService) FindByPrimaryIp(ip string) (datatypes.Virtual_Guest, bool, error) {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Finding Softlayer Virtual Guest by Primary Backend IP Address '%s' ", ip)
	instance, err := vg.softlayerClient.GetInstanceByPrimaryIpAddress(ip)
	if err != nil {
		if slErr, ok := err.(sl.Error); ok && slErr.Exception == "SoftLayer_Exception_ObjectNotFound" {
			return datatypes.Virtual_Guest{}, false, nil
		}

		return datatypes.Virtual_Guest{}, false, bosherr.WrapErrorf(err, "Failed to find Softlayer VirtualGuest with ip '%s'", ip)
	}

	return instance, true, nil
}
