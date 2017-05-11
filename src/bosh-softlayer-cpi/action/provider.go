package action

import (
	bsl "bosh-softlayer-cpi/softlayer/client"

	. "bosh-softlayer-cpi/softlayer/common"

	slvm "bosh-softlayer-cpi/softlayer/vm"
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

func NewCreatorProvider(softLayerClient bsl.Client, options ConcreteFactoryOptions, logger boshlog.Logger) CreatorProvider {
	agentEnvServiceFactory := NewSoftLayerAgentEnvServiceFactory(options.Registry, logger)

	vmFinder := slvm.NewSoftLayerFinder(
		softLayerClient,
		agentEnvServiceFactory,
		logger,
	)

	virtualGuestCreator := slvm.NewSoftLayerCreator(
		vmFinder,
		softLayerClient,
		options.Agent,
		options.Softlayer.FeatureOptions,
		options.Registry,
		logger,
	)

	return creatorProvider{
		creators: map[string]VMCreator{
			"virtualguest": virtualGuestCreator,
		},
	}
}

func (p creatorProvider) Get(name string) VMCreator {
	return p.creators[name]
}

func NewDeleterProvider(softLayerClient bsl.Client, logger boshlog.Logger, vmFinder VMFinder) DeleterProvider {
	virtualGuestDeleter := slvm.NewSoftLayerVMDeleter(
		softLayerClient,
		logger,
		vmFinder,
	)

	return deleterProvider{
		deleters: map[string]VMDeleter{
			"virtualguest": virtualGuestDeleter,
		},
	}
}

func (p deleterProvider) Get(name string) VMDeleter {
	return p.deleters[name]
}
