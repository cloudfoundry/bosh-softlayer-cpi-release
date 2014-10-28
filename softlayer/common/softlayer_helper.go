package common

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	sl "github.com/maximilien/softlayer-go/softlayer"
)

var (
	TIMEOUT          time.Duration
	POLLING_INTERVAL time.Duration
)

func ConfigureMetadataOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int, metadata string, timeout, pollingInterval time.Duration) error {
	err := WaitForVirtualGuest(softLayerClient, virtualGuestId, "RUNNING", timeout, pollingInterval)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d`", virtualGuestId))
	}

	err = WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient, virtualGuestId, timeout, pollingInterval)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d` to have no pending transactions", virtualGuestId))
	}

	err = SetMetadataOnVirtualGuest(softLayerClient, virtualGuestId, metadata)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Setting metadata on VirtualGuest `%d`", virtualGuestId))
	}

	err = ConfigureMetadataDiskOnVirtualGuest(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata disk on VirtualGuest `%d`", virtualGuestId))
	}

	err = WaitForVirtualGuest(softLayerClient, virtualGuestId, "RUNNING", timeout, pollingInterval)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d`", virtualGuestId))
	}

	return nil
}

func WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient sl.Client, virtualGuestId int, timeout, pollingInterval time.Duration) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < timeout {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
		}

		if len(activeTransactions) == 0 {
			return nil
		}
		totalTime += pollingInterval
		time.Sleep(pollingInterval)
	}

	return bosherr.New(fmt.Sprintf("Waiting for virtual guest with ID '%d' to have no active transactions", virtualGuestId))
}

func WaitForVirtualGuest(softLayerClient sl.Client, virtualGuestId int, targetState string, timeout, pollingInterval time.Duration) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < timeout {
		vgPowerState, err := virtualGuestService.GetPowerState(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, fmt.Sprintf("Getting power state for virtual guest with ID: '%d' from SoftLayer client", virtualGuestId))
		}

		if vgPowerState.KeyName == targetState {
			return nil
		}

		totalTime += pollingInterval
		time.Sleep(pollingInterval)
	}

	return bosherr.New(fmt.Sprintf("Waiting for virtual guest with ID '%d' to have be in state '%s'", virtualGuestId, targetState))
}

func SetMetadataOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int, metadata string) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	success, err := virtualGuestService.SetMetadata(virtualGuestId, metadata)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Setting metadata on VirtualGuest `%d`", virtualGuestId))
	}

	if !success {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to set metadata on VirtualGuest `%d`", virtualGuestId))
	}

	return nil
}

func ConfigureMetadataDiskOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	_, err = virtualGuestService.ConfigureMetadataDisk(virtualGuestId)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuestId))
	}

	return nil
}
