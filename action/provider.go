package action

import (
	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
	slpool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool"
	slvmpool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client"
	bosherror "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type CreatorProvider interface {
	Get(name string) (bslcvm.VMCreator, error)
}

type DeleterProvider interface {
	Get(name string) (bslcvm.VMDeleter, error)
}

type creatorProvider struct {
	creators map[string]bslcvm.VMCreator
}

type deleterProvider struct {
	deleters map[string]bslcvm.VMDeleter
}

func NewCreatorProvider(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, softLayerPoolClient *slvmpool.SoftLayerVMPool ,options ConcreteFactoryOptions, logger boshlog.Logger) CreatorProvider {

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
		options.Agent,
		logger,
		options.Softlayer.FeatureOptions,
	)

	baremetalCreator := bslcvm.NewBaremetalCreator(
		vmFinder,
		softLayerClient,
		baremetalClient,
		options.Agent,
		logger,
	)

	poolCreator := slpool.NewSoftLayerPoolCreator(
		vmFinder,
		softLayerPoolClient,
		softLayerClient,
		options.Agent,
		logger,
		options.Softlayer.FeatureOptions,
	)

	return creatorProvider{
		creators: map[string]bslcvm.VMCreator{
			"virtualguest": virtualGuestCreator,
			"baremetal":    baremetalCreator,
			"pool":		poolCreator,
		},
	}
}

func (p creatorProvider) Get(name string) (bslcvm.VMCreator, error) {
	creator, found := p.creators[name]
	if !found {
		return nil, bosherror.Errorf("Creator %s could not be found", name)
	}
	return creator, nil
}

func NewDeleterProvider(softLayerClient sl.Client, softLayerPoolClient *slvmpool.SoftLayerVMPool, logger boshlog.Logger) DeleterProvider {
	virtualGuestDeleter := bslcvm.NewSoftLayerVMDeleter(
		softLayerClient,
		logger,
	)

	poolDeleter := slpool.NewSoftLayerPoolDeleter(
		softLayerPoolClient,
		softLayerClient,
		logger,
	)

	return deleterProvider{
		deleters: map[string]bslcvm.VMDeleter{
			"virtualguest": virtualGuestDeleter,
			"pool":                poolDeleter,
		},
	}
}

func (p deleterProvider) Get(name string) (bslcvm.VMDeleter, error) {
	deleter, found := p.deleters[name]
	if !found {
		return nil, bosherror.Errorf("Deleter %s could not be found", name)
	}
	return deleter, nil
}
