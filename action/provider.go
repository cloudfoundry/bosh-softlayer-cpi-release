package action

import (
	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
	bosherror "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Provider interface {
	Get(name string) (bslcvm.VMCreator, error)
}

type provider struct {
	creators map[string]bslcvm.VMCreator
}

func NewProvider(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, options ConcreteFactoryOptions, logger boshlog.Logger) Provider {

	agentEnvServiceFactory := bslcvm.NewSoftLayerAgentEnvServiceFactory(options.AgentEnvService, options.Registry, logger)

	vmFinder := bslcvm.NewSoftLayerFinder(
		softLayerClient,
		baremetalClient,
		agentEnvServiceFactory,
		logger,
	)

	virtualGuestCreator := bslcvm.NewSoftLayerCreator(
		vmFinder,
		softLayerClient,
		agentEnvServiceFactory,
		options.Agent,
		logger,
	)

	baremetalCreator := bslcvm.NewBaremetalCreator(
		vmFinder,
		softLayerClient,
		baremetalClient,
		agentEnvServiceFactory,
		options.Agent,
		logger,
	)

	return provider{
		creators: map[string]bslcvm.VMCreator{
			"virtualguest": virtualGuestCreator,
			"baremetal":    baremetalCreator,
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
