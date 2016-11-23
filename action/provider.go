package action

import (
	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	sl "github.com/maximilien/softlayer-go/softlayer"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	slhw "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/hardware"
	slpool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool"
	operations "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm"
	slvm "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

//go:generate counterfeiter -o fakes/fake_creator_provider.go . CreatorProvider
type CreatorProvider interface {
	Get(name string) VMCreator
}

//go:generate counterfeiter -o fakes/fake_deleter_provider.go . DeleterProvider
type DeleterProvider interface {
	Get(name string) VMDeleter
}

type creatorProvider struct {
	creators map[string]VMCreator
}

type deleterProvider struct {
	deleters map[string]VMDeleter
}

func NewCreatorProvider(softLayerClient sl.Client, baremetalClient bmscl.BmpClient, softLayerPoolClient operations.SoftLayerPoolClient, options ConcreteFactoryOptions, logger boshlog.Logger) CreatorProvider {
	agentEnvServiceFactory := NewSoftLayerAgentEnvServiceFactory(options.AgentEnvService, options.Registry, logger)

	vmFinder := slvm.NewSoftLayerFinder(
		softLayerClient,
		baremetalClient,
		agentEnvServiceFactory,
		logger,
	)

	virtualGuestCreator := slvm.NewSoftLayerCreator(
		vmFinder,
		softLayerClient,
		options.Agent,
		logger,
		options.Softlayer.FeatureOptions,
	)

	baremetalCreator := slhw.NewBaremetalCreator(
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
		creators: map[string]VMCreator{
			"virtualguest": virtualGuestCreator,
			"baremetal":    baremetalCreator,
			"pool":         poolCreator,
		},
	}
}

func (p creatorProvider) Get(name string) VMCreator {
	return p.creators[name]
}

func NewDeleterProvider(softLayerClient sl.Client, softLayerPoolClient operations.SoftLayerPoolClient, logger boshlog.Logger) DeleterProvider {
	virtualGuestDeleter := slvm.NewSoftLayerVMDeleter(
		softLayerClient,
		logger,
	)

	poolDeleter := slpool.NewSoftLayerPoolDeleter(
		softLayerPoolClient,
		softLayerClient,
		logger,
	)

	return deleterProvider{
		deleters: map[string]VMDeleter{
			"virtualguest": virtualGuestDeleter,
			"pool":         poolDeleter,
		},
	}
}

func (p deleterProvider) Get(name string) VMDeleter {
	return p.deleters[name]
}
