package registry

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type SoftlayerUserData struct {
	Server   SoftlayerUserDataServerName       `json:"server"`
	Registry SoftlayerUserDataRegistryEndpoint `json:"registry"`
	DNS      SoftlayerUserDataDNSItems         `json:"dns,omitempty"`
}

type SoftlayerUserDataServerName struct {
	Name string `json:"name"`
}

type SoftlayerUserDataRegistryEndpoint struct {
	Endpoint string `json:"endpoint"`
}

type SoftlayerUserDataDNSItems struct {
	NameServer []string `json:"nameserver,omitempty"`
}

// ClientOptions are the options used to create a BOSH Registry client.
type ClientOptions struct {
	// BOSH Registry protocol
	Protocol string `json:"protocol,omitempty"`

	// BOSH Registry hostname
	Host string `json:"address,omitempty"`

	// BOSH Registry port
	Port int `json:"port,omitempty"`

	// BOSH Registry username
	Username string `json:"user,omitempty"`

	// BOSH Registry password
	Password string `json:"password,omitempty"`

	// BOSH Registry TLS options (only when using protocol https)
	TLS ClientTLSOptions `json:"tls,omitempty"`

	Endpoint string `json:"endpoint,omitempty"`

	HTTPOptions HttpRegistryOptions `json:"http,omitempty"`
}

type HttpRegistryOptions struct {
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// ClientTLSOptions are the TLS options used to create a BOSH Registry client.
type ClientTLSOptions struct {
	// If the Client must skip the verification of the server certificates
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`

	// Certificate file (PEM format)
	CertFile string `json:"certfile,omitempty"`

	// Private key file (PEM format)
	KeyFile string `json:"keyfile,omitempty"`

	// Roor CA certificate file (PEM format)
	CACertFile string `json:"cacertfile,omitempty"`
}

//// Endpoint returns the BOSH Registry endpoint.
//func (o ClientOptions) Endpoint() string {
//	return fmt.Sprintf("%s://%s:%d", "http", o.Host, o.Port)
//}

// EndpointWithCredentials returns the BOSH Registry endpoint including credentials.
func (o ClientOptions) EndpointWithCredentials() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d", "http", o.HTTPOptions.User, o.HTTPOptions.Password, o.Host, o.HTTPOptions.Port)
}

// Validate validates the Client options.
func (o ClientOptions) Validate() error {
	if o.Protocol == "" {
		return bosherr.Error("Must provide a non-empty Protocol")
	}

	if o.Host == "" {
		return bosherr.Error("Must provide a non-empty Host")
	}

	if o.Port == 0 {
		return bosherr.Error("Must provide a non-empty Port")
	}

	if o.Username == "" {
		return bosherr.Error("Must provide a non-empty Username")
	}

	if o.Password == "" {
		return bosherr.Error("Must provide a non-empty Password")
	}

	if o.Protocol == "https" {
		err := o.TLS.Validate()
		if err != nil {
			return bosherr.WrapError(err, "Validating TLS configuration")
		}
	}

	return nil
}

// Validate validates the TLS options.
func (o ClientTLSOptions) Validate() error {
	if o.CertFile == "" {
		return bosherr.Error("Must provide a non-empty CertFile")
	}

	if o.KeyFile == "" {
		return bosherr.Error("Must provide a non-empty KeyFile")
	}

	return nil
}
