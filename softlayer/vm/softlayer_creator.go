package vm

import (
	bosherr "bosh/errors"
	boshlog "bosh/logger"
	boshuuid "bosh/uuid"
	wrdn "github.com/cloudfoundry-incubator/garden/warden"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

const softLayerCreatorLogTag = "SoftLayerCreator"

type SoftLayerCreator struct {
	uuidGen boshuuid.Generator

	softLayerClient        wrdn.Client
	agentEnvServiceFactory AgentEnvServiceFactory

	hostBindMounts  HostBindMounts
	guestBindMounts GuestBindMounts

	agentOptions AgentOptions
	logger       boshlog.Logger
}

func NewSoftLayerCreator(
	uuidGen boshuuid.Generator,
	softLayerClient wrdn.Client,
	agentEnvServiceFactory AgentEnvServiceFactory,
	hostBindMounts HostBindMounts,
	guestBindMounts GuestBindMounts,
	agentOptions AgentOptions,
	logger boshlog.Logger,
) SoftLayerCreator {
	return SoftLayerCreator{
		uuidGen: uuidGen,

		softLayerClient:        softLayerClient,
		agentEnvServiceFactory: agentEnvServiceFactory,

		hostBindMounts:  hostBindMounts,
		guestBindMounts: guestBindMounts,

		agentOptions: agentOptions,
		logger:       logger,
	}
}

func (c SoftLayerCreator) Create(agentID string, stemcell bslcstem.Stemcell, networks Networks, env Environment) (VM, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Generating VM id")
	}

	networkIP, err := c.resolveNetworkIP(networks)
	if err != nil {
		return SoftLayerVM{}, err
	}

	hostEphemeralBindMountPath, hostPersistentBindMountsDir, err := c.makeHostBindMounts(id)
	if err != nil {
		return SoftLayerVM{}, err
	}

	containerSpec := wrdn.ContainerSpec{
		Handle:     id,
		RootFSPath: stemcell.DirPath(),
		Network:    networkIP,
		BindMounts: []wrdn.BindMount{
			wrdn.BindMount{
				SrcPath: hostEphemeralBindMountPath,
				DstPath: c.guestBindMounts.MakeEphemeral(),
				Mode:    wrdn.BindMountModeRW,
				Origin:  wrdn.BindMountOriginHost,
			},
			wrdn.BindMount{
				SrcPath: hostPersistentBindMountsDir,
				DstPath: c.guestBindMounts.MakePersistent(),
				Mode:    wrdn.BindMountModeRW,
				Origin:  wrdn.BindMountOriginHost,
			},
		},
		Properties: wrdn.Properties{},
	}

	c.logger.Debug(softLayerCreatorLogTag, "Creating container with spec %#v", containerSpec)

	container, err := c.softLayerClient.Create(containerSpec)
	if err != nil {
		return SoftLayerVM{}, bosherr.WrapError(err, "Creating container")
	}

	agentEnv := NewAgentEnvForVM(agentID, id, networks, env, c.agentOptions)

	agentEnvService := c.agentEnvServiceFactory.New(container)

	err = agentEnvService.Update(agentEnv)
	if err != nil {
		c.cleanUpContainer(container)
		return SoftLayerVM{}, bosherr.WrapError(err, "Updating container's agent env")
	}

	err = c.startAgentInContainer(container)
	if err != nil {
		c.cleanUpContainer(container)
		return SoftLayerVM{}, err
	}

	vm := NewSoftLayerVM(
		id,
		c.softLayerClient,
		agentEnvService,
		c.hostBindMounts,
		c.guestBindMounts,
		c.logger,
	)

	return vm, nil
}

func (c SoftLayerCreator) resolveNetworkIP(networks Networks) (string, error) {
	var network Network

	switch len(networks) {
	case 0:
		return "", bosherr.New("Expected exactly one network; received zero")
	case 1:
		network = networks.First()
	default:
		return "", bosherr.New("Expected exactly one network; received multiple")
	}

	if network.IsDynamic() {
		return "", nil
	}

	return network.IP, nil
}

func (c SoftLayerCreator) makeHostBindMounts(id string) (string, string, error) {
	ephemeralBindMountPath, err := c.hostBindMounts.MakeEphemeral(id)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Making host ephemeral bind mount path")
	}

	persistentBindMountsDir, err := c.hostBindMounts.MakePersistent(id)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Making host persistent bind mounts dir")
	}

	return ephemeralBindMountPath, persistentBindMountsDir, nil
}

func (c SoftLayerCreator) startAgentInContainer(container wrdn.Container) error {
	processSpec := wrdn.ProcessSpec{
		Path:       "/usr/sbin/runsvdir-start",
		Privileged: true,
	}

	// Do not Wait() for the process to finish
	_, err := container.Run(processSpec, wrdn.ProcessIO{})
	if err != nil {
		return bosherr.WrapError(err, "Running BOSH Agent in container")
	}

	return nil
}

func (c SoftLayerCreator) cleanUpContainer(container wrdn.Container) {
	// false is to kill immediately
	err := container.Stop(false)
	if err != nil {
		c.logger.Error(softLayerCreatorLogTag, "Failed destroying container '%s': %s", container.Handle, err.Error())
	}
}
