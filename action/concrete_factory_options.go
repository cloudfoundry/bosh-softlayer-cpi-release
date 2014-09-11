package action

import (
	bosherr "bosh/errors"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

type ConcreteFactoryOptions struct {
	StemcellsDir string
	DisksDir     string

	HostEphemeralBindMountsDir  string // e.g. /var/vcap/store/ephemeral_disks
	HostPersistentBindMountsDir string // e.g. /var/vcap/store/persistent_disks

	GuestEphemeralBindMountPath  string // e.g. /var/vcap/data
	GuestPersistentBindMountsDir string // e.g. /softlayer-cpi-dev

	Agent bslcvm.AgentOptions
}

func (o ConcreteFactoryOptions) Validate() error {
	if o.StemcellsDir == "" {
		return bosherr.New("Must provide non-empty StemcellsDir")
	}

	if o.DisksDir == "" {
		return bosherr.New("Must provide non-empty DisksDir")
	}

	if o.HostEphemeralBindMountsDir == "" {
		return bosherr.New("Must provide non-empty HostEphemeralBindMountsDir")
	}

	if o.HostPersistentBindMountsDir == "" {
		return bosherr.New("Must provide non-empty HostPersistentBindMountsDir")
	}

	if o.GuestEphemeralBindMountPath == "" {
		return bosherr.New("Must provide non-empty GuestEphemeralBindMountPath")
	}

	if o.GuestPersistentBindMountsDir == "" {
		return bosherr.New("Must provide non-empty GuestPersistentBindMountsDir")
	}

	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	return nil
}
