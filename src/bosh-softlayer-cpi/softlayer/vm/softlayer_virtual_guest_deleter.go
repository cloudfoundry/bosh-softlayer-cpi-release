package vm

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "bosh-softlayer-cpi/softlayer/common"

	slhelper "bosh-softlayer-cpi/softlayer/common/helpers"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const SOFTLAYER_VM_DELETER_LOG_TAG = "SoftLayerVMDeleter"

type softLayerVMDeleter struct {
	softLayerClient sl.Client
	logger          boshlog.Logger
	vmFinder        VMFinder
}

func NewSoftLayerVMDeleter(softLayerClient sl.Client, logger boshlog.Logger, vmFinder VMFinder) VMDeleter {
	return &softLayerVMDeleter{
		softLayerClient: softLayerClient,
		logger:          logger,
		vmFinder:        vmFinder,
	}
}

func (c *softLayerVMDeleter) Delete(cid int) error {
	virtualGuestService, err := c.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	err = slhelper.WaitForVirtualGuestToHaveNoRunningTransactions(c.softLayerClient, cid)
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d` to have no pending transactions before deleting vm", cid))
		}
	}

	vm, found, err := c.vmFinder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding VirtualGuest with id: %d.", cid)
	} else {
		if !found {
			return bosherr.WrapErrorf(err, "Cannot find VirtualGuest with id: %d.", cid)
		}
	}

	err = vm.DeleteAgentEnv()
	if err != nil {
		return bosherr.WrapError(err, "Deleting VM's agent env")
	}

	_, err = virtualGuestService.DeleteObject(cid)
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return bosherr.WrapError(err, "Deleting SoftLayer VirtualGuest from client")
		}
	}

	return nil
}
