package action

import (
	"os"
	"strconv"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	bm "github.com/maximilien/bosh-softlayer-cpi/softlayer/baremetal"
)

const (
	MEMORY_CON     = "SL_BARE_METAL_MEMORY"
	PROCESSOR_CON  = "SL_BARE_METAL_PROCESSOR"
	DISK_CON       = "SL_BARE_METAL_DISK"
	HOST_CON       = "SL_BARE_METAL_HOST"
	DOMAIN_CON     = "SL_BARE_METAL_DOMAIN"
	OS_CON         = "SL_BARE_METAL_OS"
	DATACENTER_CON = "SL_DATA_CENTER"

	MAX_RETRIES = 120
)

type EstablishBareMetalEnv struct {
	bmCreator bm.BaremetalCreator
	bmFinder  bm.BaremetalFinder
}

func NewEstablishBareMetalEnv(bmCreator bm.BaremetalCreator, bmFinder bm.BaremetalFinder) EstablishBareMetalEnv {
	return EstablishBareMetalEnv{
		bmCreator: bmCreator,
		bmFinder:  bmFinder,
	}
}

func (b EstablishBareMetalEnv) Run() (interface{}, error) {

	// Step #1, use baremetal creator to create a new baremetal server
	memoryS := os.Getenv(MEMORY_CON)
	processorS := os.Getenv(PROCESSOR_CON)
	disksizeS := os.Getenv(DISK_CON)
	host := os.Getenv(HOST_CON)
	domain := os.Getenv(DOMAIN_CON)
	ostype := os.Getenv(OS_CON)
	datacenter := os.Getenv(DATACENTER_CON)

	memory, err := strconv.Atoi(memoryS)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Invalid baremetal creation memory parameter: %s", memoryS)
	}

	processor, err := strconv.Atoi(processorS)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Invalid baremetal creation processor parameter: %s", processorS)
	}

	disksize, err := strconv.Atoi(disksizeS)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Invalid baremetal creation disk size parameter: %s", disksizeS)
	}

	baremetal, err := b.bmCreator.Create(memory, processor, disksize, host, domain, ostype, datacenter)
	if err != nil {
		return nil, bosherr.WrapError(err, "Create baremetal server error")
	}

	for step := 1; baremetal.ProvisionDate == nil || step >= MAX_RETRIES; step++ {
		time.Sleep(60 * time.Second)

		baremetal, err = b.bmFinder.Find(baremetal.GlobalIdentifier)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Cannot find the baremetal server of id: %s", baremetal.GlobalIdentifier)
		}
	}

	if baremetal.ProvisionDate == nil {
		return nil, bosherr.WrapError(err, "Baremetal server creation timeout")
	}

	return nil, nil
}
