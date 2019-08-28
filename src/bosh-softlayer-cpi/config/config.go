package config

import (
	"encoding/json"
	"regexp"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"bosh-softlayer-cpi/logger"
	"bosh-softlayer-cpi/registry"
	boslconfig "bosh-softlayer-cpi/softlayer/config"
)

type Config struct {
	Cloud Cloud `json:"cloud"`
}

type Cloud struct {
	Plugin     string        `json:"plugin"`
	Properties CPIProperties `json:"properties"`
}

type CPIProperties struct {
	SoftLayer boslconfig.Config
	Agent     registry.AgentOptions
	Registry  registry.ClientOptions
}

func NewConfigFromPath(configFile string, fs boshsys.FileSystem, logger logger.Logger) (Config, error) {
	var config Config

	if configFile == "" {
		return config, bosherr.Errorf("Must provide a config file")
	}

	logger.Debug("File System", "Reading file %s", configFile)

	bytes, err := fs.ReadFileWithOpts(configFile, boshsys.ReadOpts{Quiet: true})
	if err != nil {
		return config, bosherr.WrapErrorf(err, "Reading config file '%s'", configFile)
	}

	logger.DebugWithDetails("File System", "Read content", string(redact(bytes)))

	if err = json.Unmarshal(bytes, &config); err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config contents")
	}

	if err = config.Validate(); err != nil {
		return config, bosherr.WrapError(err, "Validating config")
	}

	return config, nil
}

func NewConfigFromString(configString string) (Config, error) {
	var config Config
	var err error
	if configString == "" {
		return config, bosherr.Errorf("Must provide a config")
	}

	if err = json.Unmarshal([]byte(configString), &config); err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling config contents")
	}

	if err = config.Validate(); err != nil {
		return config, bosherr.WrapError(err, "Validating config")
	}

	return config, nil
}

func (c Config) Validate() error {
	if c.Cloud.Plugin != "softlayer" {
		return bosherr.Errorf("Unsupported cloud plugin type %q", c.Cloud.Plugin)
	}
	if err := c.Cloud.Properties.SoftLayer.Validate(); err != nil {
		return bosherr.WrapError(err, "Validating SoftLayer configuration")
	}
	if err := c.Cloud.Properties.Agent.Validate(); err != nil {
		return bosherr.WrapError(err, "Validating agent configuration")
	}
	//if err := c.Cloud.Properties.Registry.Validate(); err != nil {
	//	return bosherr.WrapError(err, "Validating registry configuration")
	//}

	return nil
}

func redact(bs []byte) []byte {
	s := string(bs)

	hiddenStr1 := "\"password\":\"<redact>\""      // replacement string
	r1 := regexp.MustCompile(`"password":"[^"]*"`) // original string
	s1 := r1.ReplaceAllString(s, hiddenStr1)

	hiddenStr2 := "\"api_key\":\"<redact>\""
	r2 := regexp.MustCompile(`"api_key":"[^"]*"`)
	s2 := r2.ReplaceAllString(s1, hiddenStr2)

	hiddenStr3 := "//registry:<redact>"
	r3 := regexp.MustCompile(`//registry:[^@]*`)
	s3 := r3.ReplaceAllString(s2, hiddenStr3)

	hiddenStr4 := "//nats:<redact>"
	r4 := regexp.MustCompile(`//nats:[^@]*`)
	s4 := r4.ReplaceAllString(s3, hiddenStr4)

	return []byte(s4)
}
