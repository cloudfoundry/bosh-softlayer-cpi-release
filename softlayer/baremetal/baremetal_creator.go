package baremetal

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
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

	err := c.validate_arguments(memory, processor, disksize, host, domain, ostype, datacenter)
	if err != nil {
		return datatypes.SoftLayer_Hardware{}, err
	}

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

func (c BaremetalCreator) validate_arguments(memory int, processor int, disksize int, host string, domain string, ostype string, datacenter string) error {

	if memory <= 0 {
		return bosherr.Errorf("memory can not be negative: %d", memory)
	}

	if processor <= 0 {
		return bosherr.Errorf("processor can not be negative: %d", processor)
	}

	if disksize <= 0 {
		return bosherr.Errorf("disk size can not be negative: %d", disksize)
	}

	if host == "" {
		return bosherr.Error("host can not be empty.")
	}

	if domain == "" {
		return bosherr.Error("domain can not be empty.")
	}

	if ostype == "" {
		return bosherr.Error("os type can not be empty.")
	}

	if datacenter == "" {
		return bosherr.Error("data center can not be empty.")
	}

	return nil
}
