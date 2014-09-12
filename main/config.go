package main

import (
	"encoding/json"

	bosherr "bosh/errors"
	boshsys "bosh/system"

	bslcaction "github.com/maximilien/bosh-softlayer-cpi/action"
)

type Config struct {
	SoftLayer SoftLayerConfig

	Actions bslcaction.ConcreteFactoryOptions
}

type SoftLayerConfig struct {
	Username string
	ApiKey string
}

func NewConfigFromPath(path string, fs boshsys.FileSystem) (Config, error) {
	var config Config

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return config, bosherr.WrapError(err, "Reading config %s", path)
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
	err := c.SoftLayer.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating SoftLayer configuration")
	}

	err = c.Actions.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Actions configuration")
	}

	return nil
}

func (c SoftLayerConfig) Validate() error {
	if c.Username == "" {
		return bosherr.New("Must provide non-empty Username")
	}

	if c.ApiKey == "" {
		return bosherr.New("Must provide non-empty ApiKey")
	}

	return nil
}
