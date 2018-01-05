package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Username             string `json:"username"`
	ApiKey               string `json:"api_key"`
	ApiEndpoint          string `json:"api_endpoint"`
	DisableOsReload      bool   `json:"disable_os_reload"`
	PublicKey            string `json:"ssh_public_key"`
	PublicKeyFingerPrint string `json:"ssh_public_key_fingerprint"`
	EnableVps            bool   `json:"enable_vps"`
	Trace                bool   `json:"trace"`
	VpsHost              string `json:"vps_host"`
	VpsPort              int    `json:"vps_port"`
	SwiftUsername        string `json:"swift_username"`
	SwiftEndpoint        string `json:"swift_endpoint"`
	// SWIFT password is also SoftLayer API key
}

func (c Config) Validate() error {
	if c.Username == "" {
		return bosherr.Error("Must provide non-empty Username")
	}

	if c.ApiKey == "" {
		return bosherr.Error("Must provide non-empty ApiKey")
	}

	return nil
}
