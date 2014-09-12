package vm

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type AgentOptions struct {
	// e.g. "https://user:password@127.0.0.1:4321/agent"
	Mbus string

	// e.g. ["0.us.pool.ntp.org"]. Ok to be empty
	NTP []string

	Blobstore BlobstoreOptions
}

type BlobstoreOptions struct {
	// e.g. local
	Type string

	Options map[string]interface{}
}

func (o AgentOptions) Validate() error {
	if o.Mbus == "" {
		return bosherr.New("Must provide non-empty Mbus")
	}

	err := o.Blobstore.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Blobstore configuration")
	}

	return nil
}

func (o BlobstoreOptions) Validate() error {
	if o.Type == "" {
		return bosherr.New("Must provide non-empty Type")
	}

	return nil
}
