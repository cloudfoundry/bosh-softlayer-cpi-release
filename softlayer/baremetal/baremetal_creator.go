package baremetal

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

const bmCreatorLogTag = "BaremetalCreator"

type BaremetalCreator struct {
	client sl.Client
	logger boshlog.Logger
}

func NewBaremetalCreator(
	client sl.Client,
	logger boshlog.Logger,
) BaremetalCreator {
	return BaremetalCreator{
		client: client,
		logger: logger,
	}
}

func (c BaremetalCreator) Create(memory int, processor int, disksize int, host string, domain string, ostype string, datacenter string) (datatypes.SoftLayer_Hardware, error) {
	c.logger.Debug(bmCreatorLogTag, "Creating baremetal %s.%s of %s with memory: %d, processors: %d, disk size '%d'", host, domain, ostype, memory, processor, disksize)

	service, err := c.client.GetSoftLayer_Hardware_Service()
	if err != nil {
		return datatypes.SoftLayer_Hardware{}, bosherr.WrapError(err, "Get hardware service error")
	}

	template := datatypes.SoftLayer_Hardware_Template{
		Hostname:                     host,
		Domain:                       domain,
		ProcessorCoreAmount:          processor,
		MemoryCapacity:               memory,
		HourlyBillingFlag:            true,
		OperatingSystemReferenceCode: ostype,

		Datacenter: &datatypes.Datacenter{
			Name: datacenter,
		},
	}

	baremetal, err := service.CreateObject(template)

	if err != nil {
		return datatypes.SoftLayer_Hardware{}, bosherr.WrapError(err, "Create baremetal error")
	}

	return baremetal, nil
}
