package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type AgentOptions struct {
	// e.g. "https://user:password@127.0.0.1:4321/agent"
	Mbus string

	// e.g. ["0.us.pool.ntp.org"]. Ok to be empty
	NTP []string

	Blobstore BlobstoreOptions

	//vcap password
	VcapPassword string
}

type RegistryOptions struct {
	Host     string
	Port     int
	Username string
	Password string
}

type BlobstoreOptions struct {
	// e.g. local, dav
	Type string

	Options map[string]interface{}
}

func (o AgentOptions) Validate() error {
	if o.Mbus == "" {
		return bosherr.Error("Must provide non-empty Mbus")
	}

	err := o.Blobstore.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Blobstore configuration")
	}

	return nil
}

func (o BlobstoreOptions) Validate() error {
	if o.Type == "" {
		return bosherr.Error("Must provide non-empty Type")
	}

	return nil
}
