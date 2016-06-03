package action

import (
	sl "github.com/maximilien/softlayer-go/softlayer"
	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"

	bosherror "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
)

type Provider interface {
	Get(name string) (bslcvm.VMCreator, error)
}

type provider struct {
	creators map[string]bslcvm.VMCreator
}

func NewProvider(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, options ConcreteFactoryOptions, logger boshlog.Logger, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem) Provider {

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(options., options.Registry, logger)

	vmFinder := bslcvm.NewSoftLayerFinder(
		softLayerClient,
		agentEnvServiceFactory,
		logger,
		uuidGenerator,
		fs,
	)

	virtualGuestCreator := bslcvm.NewSoftLayerCreator(
		softLayerClient,
		agentEnvServiceFactory,
		options.Agent,
		logger,
		uuidGenerator,
		fs,
		vmFinder,
	)

	baremetalCreator := bslcvm.NewBaremetalCreator(
		softLayerClient,
		baremetalClient,
		agentEnvServiceFactory,
		options.Agent,
		logger,
		uuidGenerator,
		fs,
		vmFinder,
	)

	return provider{
		creators: map[string]bslcvm.VMCreator{
			"virtualguest": virtualGuestCreator,
			"baremetal": baremetalCreator,
		},
	}
}

func (p provider) Get(name string) (bslcvm.VMCreator, error) {
	creator, found := p.creators[name]
	if !found {
		return nil, bosherror.Errorf("Creator %s could not be found", name)
	}
	return creator, nil
}