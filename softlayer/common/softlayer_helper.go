package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	"github.com/pivotal-golang/clock"

	slcommon "github.com/maximilien/softlayer-go/common"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

var (
	TIMEOUT          time.Duration
	POLLING_INTERVAL time.Duration
)

type SoftLayer_Hardware_Parameters struct {
	Parameters []datatypes.SoftLayer_Hardware `json:"parameters"`
}

func IscsiHasAllowedHardware(softLayerClient sl.Client, volumeId int, hardwareId int) (bool, error) {
	filter := string(`{"allowedHardwares":{"id":{"operation":"` + strconv.Itoa(hardwareId) + `"}}}`)
	response, errorCode, err := softLayerClient.GetHttpClient().DoRawHttpRequestWithObjectFilterAndObjectMask(fmt.Sprintf("%s/%d/getAllowedVirtualGuests.json", "SoftLayer_Network_Storage", volumeId), []string{"id"}, fmt.Sprintf(string(filter)), "GET", new(bytes.Buffer))

	if err != nil {
		return false, errors.New(fmt.Sprintf("Cannot check authentication for volume %d in vm %d", volumeId, hardwareId))
	}

	if slcommon.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Network_Storage#hasAllowedVirtualGuest, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	hardware := []datatypes.SoftLayer_Hardware{}
	err = json.Unmarshal(response, &hardware)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Failed to unmarshal response of checking authentication for volume %d in vm %d", volumeId, hardwareId))
	}

	if len(hardware) > 0 {
		return true, nil
	}

	return false, nil
}

func AttachHardwareIscsiVolume(softLayerClient sl.Client, hardware datatypes.SoftLayer_Hardware, volumeId int) (bool, error) {
	parameters := SoftLayer_Hardware_Parameters{
		Parameters: []datatypes.SoftLayer_Hardware{
			hardware,
		},
	}
	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return false, err
	}

	resp, errorCode, err := softLayerClient.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/allowAccessFromHardware.json", "SoftLayer_Network_Storage", volumeId), "PUT", bytes.NewBuffer(requestBody))

	if err != nil {
		return false, err
	}

	if slcommon.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Network_Storage#attachIscsiVolume, HTTP error code: '%d'", errorCode)
		return false, errors.New(errorMessage)
	}

	allowable, err := strconv.ParseBool(string(resp[:]))
	if err != nil {
		return false, nil
	}

	return allowable, nil
}

func DetachHardwareIscsiVolume(softLayerClient sl.Client, hardware datatypes.SoftLayer_Hardware, volumeId int) error {
	parameters := SoftLayer_Hardware_Parameters{
		Parameters: []datatypes.SoftLayer_Hardware{
			hardware,
		},
	}
	requestBody, err := json.Marshal(parameters)
	if err != nil {
		return err
	}

	_, errorCode, err := softLayerClient.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/removeAccessFromVirtualGuest.json", "SoftLayer_Network_Storage", volumeId), "PUT", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	if slcommon.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Account#getAccountStatus, HTTP error code: '%d'", errorCode)
		return errors.New(errorMessage)
	}

	return nil
}

func GetHardwareAllowedHost(softLayerClient sl.Client, instanceId int) (datatypes.SoftLayer_Network_Storage_Allowed_Host, error) {
	response, errorCode, err := softLayerClient.GetHttpClient().DoRawHttpRequest(fmt.Sprintf("%s/%d/getAllowedHost.json", "SoftLayer_Hardware", instanceId), "GET", new(bytes.Buffer))
	if err != nil {
		return datatypes.SoftLayer_Network_Storage_Allowed_Host{}, err
	}

	if slcommon.IsHttpErrorCode(errorCode) {
		errorMessage := fmt.Sprintf("softlayer-go: could not SoftLayer_Hardware#getAllowedHost, HTTP error code: '%d'", errorCode)
		return datatypes.SoftLayer_Network_Storage_Allowed_Host{}, errors.New(errorMessage)
	}

	allowedHost := datatypes.SoftLayer_Network_Storage_Allowed_Host{}
	err = json.Unmarshal(response, &allowedHost)
	if err != nil {
		return datatypes.SoftLayer_Network_Storage_Allowed_Host{}, err
	}

	return allowedHost, nil
}

func AttachEphemeralDiskToVirtualGuest(softLayerClient sl.Client, virtualGuestId int, diskSize int, logger boshlog.Logger) error {
	err := WaitForVirtualGuestLastCompleteTransaction(softLayerClient, virtualGuestId, "Service Setup")
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` has Service Setup transaction complete", virtualGuestId)
	}

	err = WaitForVirtualGuestToHaveNoRunningTransactions(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to have no pending transactions", virtualGuestId)
	}

	service, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapErrorf(err, "Attaching ephemeral disk to VirtualGuest `%d`", virtualGuestId)
	}

	receipt, err := service.AttachEphemeralDisk(virtualGuestId, diskSize)
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return err
		}
	}

	if receipt.OrderId == 0 {
		return nil
	}

	err = WaitForVirtualGuestToHaveRunningTransaction(softLayerClient, virtualGuestId, logger)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` to launch transaction", virtualGuestId)
	}

	err = WaitForVirtualGuestToHaveNoRunningTransaction(softLayerClient, virtualGuestId, logger)

	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` no transcation in progress", virtualGuestId)
	}

	err = WaitForVirtualGuestUpgradeComplete(softLayerClient, virtualGuestId)
	if err != nil {
		return bosherr.WrapErrorf(err, "Waiting for VirtualGuest `%d` upgrade complete", virtualGuestId)
	}

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

	totalTime := time.Duration(0)
	for totalTime < TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting active transaction against vitrual guest %d", virtualGuestId)
		}

		if len(activeTransactions) > 0 {
			return nil
		}

		totalTime += POLLING_INTERVAL
		time.Sleep(POLLING_INTERVAL)
	}

	return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have no active transactions", virtualGuestId)
}

func WaitForVirtualGuestToHaveNoRunningTransaction(softLayerClient sl.Client, virtualGuestId int, logger boshlog.Logger) error {

	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting active transaction against vitrual guest %d", virtualGuestId)
		}

		if len(activeTransactions) == 0 {
			return nil
		}

		totalTime += POLLING_INTERVAL
		time.Sleep(POLLING_INTERVAL)
	}

	return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have no active transactions", virtualGuestId)

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

func WaitForVirtualGuestLastCompleteTransaction(softLayerClient sl.Client, virtualGuestId int, targetTransaction string) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < TIMEOUT {
		lastTransaction, err := virtualGuestService.GetLastTransaction(virtualGuestId)
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting Last Complete Transaction for virtual guest with ID '%d'", virtualGuestId)
		}

		if strings.Contains(lastTransaction.TransactionGroup.Name, targetTransaction) && strings.Contains(lastTransaction.TransactionStatus.FriendlyName, "Complete") {
			return nil
		}

		totalTime += POLLING_INTERVAL
		time.Sleep(POLLING_INTERVAL)
	}

	return bosherr.Errorf("Waiting for virtual guest with ID '%d' to have last transaction '%s'", virtualGuestId, targetTransaction)
}

func WaitForVirtualGuestIsNotPingable(softLayerClient sl.Client, virtualGuestId int, logger boshlog.Logger) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	checkPingableRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			state, err := virtualGuestService.IsPingable(virtualGuestId)
			if err != nil {
				return false, bosherr.WrapErrorf(err, "Checking pingable against vitrual guest %d", virtualGuestId)
			} else {
				if state {
					return true, bosherr.Errorf("vitrual guest %d is pingable", virtualGuestId)
				} else {
					return false, nil
				}
			}
		})

	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(TIMEOUT, POLLING_INTERVAL, checkPingableRetryable, timeService, logger)
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return bosherr.Errorf("Waiting for virtual guest with ID '%d' is not pingable", virtualGuestId)
	}

	return nil
}

func WaitForVirtualGuestIsPingable(softLayerClient sl.Client, virtualGuestId int, logger boshlog.Logger) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	checkPingableRetryable := boshretry.NewRetryable(
		func() (bool, error) {
			state, err := virtualGuestService.IsPingable(virtualGuestId)
			if err != nil {
				return false, bosherr.WrapErrorf(err, "Checking pingable against vitrual guest %d", virtualGuestId)
			} else {
				if state {
					return false, nil
				} else {
					return true, bosherr.Errorf("vitrual guest %d is not pingable", virtualGuestId)
				}
			}
		})

	timeService := clock.NewClock()
	timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(TIMEOUT, POLLING_INTERVAL, checkPingableRetryable, timeService, logger)
	err = timeoutRetryStrategy.Try()
	if err != nil {
		return bosherr.Errorf("Waiting for virtual guest with ID '%d' is not pingable", virtualGuestId)
	}
	return nil
}

func WaitForVirtualGuestUpgradeComplete(softLayerClient sl.Client, virtualGuestId int) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < TIMEOUT {
		lastTransaction, err := virtualGuestService.GetLastTransaction(virtualGuestId)
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting Last Complete Transaction for virtual guest with ID '%d'", virtualGuestId)
		}

		if strings.Contains(lastTransaction.TransactionGroup.Name, "Cloud Migrate") && strings.Contains(lastTransaction.TransactionStatus.FriendlyName, "Complete") {
			return nil
		}

		if strings.Contains(lastTransaction.TransactionGroup.Name, "Cloud Instance Upgrade") && strings.Contains(lastTransaction.TransactionStatus.FriendlyName, "Complete") {
			return nil
		}

		totalTime += POLLING_INTERVAL
		time.Sleep(POLLING_INTERVAL)
	}

	return bosherr.Errorf("Waiting for virtual guest with ID '%d' to update complete", virtualGuestId)
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
				return true, bosherr.Errorf("The PowerState of vitrual guest %d is not targetState %s", virtualGuestId, targetState)
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

func GetObjectDetailsOnHardware(softLayerClient sl.Client, hardwareId int) (datatypes.SoftLayer_Hardware, error) {
	hardwareService, err := softLayerClient.GetSoftLayer_Hardware_Service()
	if err != nil {
		return datatypes.SoftLayer_Hardware{}, bosherr.WrapError(err, "Cannot get softlayer hardeare service.")
	}
	hardware, err := hardwareService.GetObject(hardwareId)
	if err != nil {
		return datatypes.SoftLayer_Hardware{}, bosherr.WrapErrorf(err, "Cannot get hardware with id: %d", hardwareId)
	}
	return hardware, nil
}

func GetObjectDetailsOnStorage(softLayerClient sl.Client, volumeId int) (datatypes.SoftLayer_Network_Storage, error) {
	networkStorageService, err := softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Cannot get network storage service.")
	}

	volume, err := networkStorageService.GetNetworkStorage(volumeId)
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Cannot get iSCSI volume with id: %d", volumeId)
	}

	return volume, nil
}
