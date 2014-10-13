package common

import (
	"fmt"
	"time"

	gomega "github.com/onsi/gomega"

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

	gomega.Eventually(func() int {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return len(activeTransactions)
	}, timeout, pollingInterval).Should(gomega.Equal(0), "failed waiting for virtual guest to have no active transactions")

	return nil
}

func WaitForVirtualGuest(softLayerClient sl.Client, virtualGuestId int, targetState string, timeout, pollingInterval time.Duration) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	gomega.Eventually(func() string {
		vgPowerState, err := virtualGuestService.GetPowerState(virtualGuestId)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return vgPowerState.KeyName
	}, timeout, pollingInterval).Should(gomega.Equal(targetState), fmt.Sprintf("failed waiting for virtual guest to be %s", targetState))

	return nil
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
