package common

import (
	"encoding/base64"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	"github.com/pivotal-golang/clock"

	datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"
	"strings"
)

var (
	TIMEOUT          time.Duration
	POLLING_INTERVAL time.Duration
	PAUSE_TIME       time.Duration
)

func AttachEphemeralDiskToVirtualGuest(softLayerClient sl.Client, virtualGuestId int, diskSize int, logger boshlog.Logger) error {
	err := WaitForVirtualGuest(softLayerClient, virtualGuestId, "RUNNING")
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d`", virtualGuestId)
	}

	err = WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to have no pending transactions", virtualGuestId)
	}

	service, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapErrorf(err, "Attaching ephemeral disk to VirtualGuest `%d`", virtualGuestId)
	}

	err = service.AttachEphemeralDisk(virtualGuestId, diskSize)
	if err != nil {
		return bosherr.WrapErrorf(err, "Attaching ephemeral disk to VirtualGuest `%d`", virtualGuestId)
	}

	err = WaitForVirtualGuestToHaveRunningTransaction(softLayerClient, virtualGuestId, logger)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to launch transaction", virtualGuestId)
	}

	err = WaitForVirtualGuest(softLayerClient, virtualGuestId, "RUNNING")
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d`", virtualGuestId)
	}

	return nil
}

func ConfigureMetadataOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int, metadata string, logger boshlog.Logger) error {
	err := WaitForVirtualGuest(softLayerClient, virtualGuestId, "RUNNING")
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d`", virtualGuestId)
	}

	err = WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to have no pending transactions", virtualGuestId)
	}

	err = SetMetadataOnVirtualGuest(softLayerClient, virtualGuestId, metadata)
	if err != nil {
		return bosherr.WrapErrorf(err, "Setting metadata on VirtualGuest `%d`", virtualGuestId)
	}

	err = WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to have no pending transactions", virtualGuestId)
	}

	err = ConfigureMetadataDiskOnVirtualGuest(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapErrorf(err, "Configuring metadata disk on VirtualGuest `%d`", POLLING_INTERVAL)
	}

	err = WaitForVirtualGuestToHaveRunningTransaction(softLayerClient, virtualGuestId, logger)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to launch transaction", virtualGuestId)
	}

	//The transaction (configureMetadataDisk) will shut down the guest while the metadata disk is configured. Pause 2 minutes for its back.
	//time.Sleep(PAUSE_TIME)

	err = WaitForVirtualGuest(softLayerClient, virtualGuestId, "RUNNING")
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d`", virtualGuestId)
	}

	return nil
}

func WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient sl.Client, virtualGuestId int) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transaction from SoftLayer client")
		}

		if len(activeTransactions) == 0 {
			return nil
		}

		totalTime += POLLING_INTERVAL
		time.Sleep(POLLING_INTERVAL)
	}

	return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have no active transactions", virtualGuestId)
}

func WaitForVirtualGuestToHaveRunningTransaction(softLayerClient sl.Client, virtualGuestId int, logger boshlog.Logger) error {

	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	runningTransactionsRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
			if err != nil {
				return false, bosherr.WrapErrorf(err, "Getting active transaction against vitrual guest %d", virtualGuestId)
			} else {
				if len(activeTransactions) == 1 {
					return false, nil
				}
				return true, nil
			}
		})

	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(TIMEOUT, POLLING_INTERVAL, runningTransactionsRetryable, timeService, logger)
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have active transactions", virtualGuestId)
	}

	return nil
}

func WaitForVirtualGuestToHaveNoRunningTransaction(softLayerClient sl.Client, virtualGuestId int, logger boshlog.Logger) error {

	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	runningTransactionsRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
			if err != nil {
				return false, bosherr.WrapErrorf(err, "Getting active transaction against vitrual guest %d", virtualGuestId)
			} else {
				if len(activeTransactions) == 0 {
					return false, nil
				}
				return true, nil
			}
		})

	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(TIMEOUT, POLLING_INTERVAL, runningTransactionsRetryable, timeService, logger)
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have no active transactions", virtualGuestId)
	}
	return nil
}

func WaitForVirtualGuest(softLayerClient sl.Client, virtualGuestId int, targetState string) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < TIMEOUT {
		vgPowerState, err := virtualGuestService.GetPowerState(virtualGuestId)
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting Power State for virtual guest with ID '%d'", virtualGuestId)
		}

		if strings.Contains(vgPowerState.KeyName, targetState) {
			return nil
		}

		totalTime += POLLING_INTERVAL
		time.Sleep(POLLING_INTERVAL)
	}

	return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have be in state '%s'", virtualGuestId, targetState)
}

func WaitForVirtualGuestToTargetState(softLayerClient sl.Client, virtualGuestId int, targetState string, logger boshlog.Logger) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	getTargetStateRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			vgPowerState, err := virtualGuestService.GetPowerState(virtualGuestId)
			if err != nil {
				return false, bosherr.WrapErrorf(err, "Getting PowerState from vitrual guest %d", virtualGuestId)
			} else {
				if strings.Contains(vgPowerState.KeyName, targetState) {
					return false, nil
				}
				return true, nil
			}
		})

	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(TIMEOUT, POLLING_INTERVAL, getTargetStateRetryable, timeService, logger)
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have be in state '%s'", virtualGuestId, targetState)
	}

	return nil
}

func SetMetadataOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int, metadata string) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	success, err := virtualGuestService.SetMetadata(virtualGuestId, metadata)
	if err != nil {
		return bosherr.WrapErrorf(err, "Setting metadata on VirtualGuest `%d`", virtualGuestId)
	}

	if !success {
		return bosherr.WrapErrorf(err, "Failed to set metadata on VirtualGuest `%d`", virtualGuestId)
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
		return bosherr.WrapErrorf(err, "Configuring metadata on VirtualGuest `%d`", virtualGuestId)
	}

	return nil
}

func GetUserMetadataOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int) ([]byte, error) {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return []byte{}, bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	attributes, err := virtualGuestService.GetUserData(virtualGuestId)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Getting metadata on VirtualGuest `%d`", virtualGuestId)
	}

	if len(attributes) == 0 {
		return []byte{}, bosherr.WrapErrorf(err, "Failed to get metadata on VirtualGuest `%d`", virtualGuestId)
	}

	sEnc := attributes[0].Value
	sDec, err := base64.StdEncoding.DecodeString(sEnc)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Failed to decode metadata returned from virtualGuest `%d`", virtualGuestId)
	}

	return sDec, nil
}

func GetObjectDetailsOnVirtualGuest(softLayerClient sl.Client, virtualGuestId int) (datatypes.SoftLayer_Virtual_Guest, error) {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, bosherr.WrapError(err, "Cannot get softlayer virtual guest service.")
	}
	virtualGuest, err := virtualGuestService.GetObject(virtualGuestId)
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, bosherr.WrapErrorf(err, "Cannot get virtual guest with id: %d", virtualGuestId)
	}
	return virtualGuest, nil
}

func GetObjectDetailsOnStorage(softLayerClient sl.Client, volumeId int) (datatypes.SoftLayer_Network_Storage, error) {
	networkStorageService, err := softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Cannot get network storage service.")
	}

	volume, err := networkStorageService.GetIscsiVolume(volumeId)
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Cannot get iSCSI volume with id: %d", volumeId)
	}
	return volume, nil
}
