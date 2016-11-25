package pool

import (
	"fmt"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	strfmt "github.com/go-openapi/strfmt"

	slhelper "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"

	operations "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm"
	sl "github.com/maximilien/softlayer-go/softlayer"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/models"
)

type softLayerPoolDeleter struct {
	softLayerClient       sl.Client
	softLayerVmPoolClient operations.SoftLayerPoolClient
	logger                boshlog.Logger
}

func NewSoftLayerPoolDeleter(softLayerVmPoolClient operations.SoftLayerPoolClient, softLayerClient sl.Client, logger boshlog.Logger) VMDeleter {
	return &softLayerPoolDeleter{
		softLayerClient:       softLayerClient,
		softLayerVmPoolClient: softLayerVmPoolClient,
		logger:                logger,
	}
}

func (c *softLayerPoolDeleter) Delete(cid int) error {
	_, err := c.softLayerVmPoolClient.GetVMByCid(operations.NewGetVMByCidParams().WithCid(int32(cid)))
	if err != nil {
		_, ok := err.(*operations.GetVMByCidNotFound)
		if ok {
			virtualGuest, err := slhelper.GetObjectDetailsOnVirtualGuest(c.softLayerClient, cid)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Getting virtual guest %d details from SoftLayer", cid))
			}

			slPoolVm := &models.VM{
				Cid:         int32(cid),
				CPU:         int32(virtualGuest.StartCpus),
				MemoryMb:    int32(virtualGuest.MaxMemory),
				IP:          strfmt.IPv4(virtualGuest.PrimaryBackendIpAddress),
				Hostname:    virtualGuest.FullyQualifiedDomainName,
				PrivateVlan: int32(virtualGuest.PrimaryBackendNetworkComponent.NetworkVlan.Id),
				PublicVlan:  int32(virtualGuest.PrimaryNetworkComponent.NetworkVlan.Id),
				State:       models.StateFree,
			}
			_, err = c.softLayerVmPoolClient.AddVM(operations.NewAddVMParams().WithBody(slPoolVm))
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Adding vm %d to pool", cid))
			}
			return nil
		}
		return bosherr.WrapError(err, "Removing vm from pool")
	}

	free := models.VMState{
		State: models.StateFree,
	}
	_, err = c.softLayerVmPoolClient.UpdateVMWithState(operations.NewUpdateVMWithStateParams().WithBody(&free).WithCid(int32(cid)))
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating state of vm %d in pool to free", cid)
	}

	return nil
}
