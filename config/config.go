package config

import (
	"encoding/json"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	bslcaction "github.com/cloudfoundry/bosh-softlayer-cpi/action"
)

type Config struct {
	Cloud CloudConfig `json:"cloud"`
}

type CloudConfig struct {
	Plugin     string                            `json:"plugin"`
	Properties bslcaction.ConcreteFactoryOptions `json:"properties"`
}

func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapErrorf(err, "Reading config %s", path)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config")
	}

	err = config.Validate()
	if err != nil {
		return config, bosherr.WrapError(err, "Validating config")
	}

	return config, nil
}

func (c Config) Validate() error {
	if !strings.EqualFold(c.Cloud.Plugin, "softlayer") {
		return bosherr.Error("Should softlayer plugin")
	}

	err := c.Cloud.Properties.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Cloud Properties")
	}

	return nil
}
