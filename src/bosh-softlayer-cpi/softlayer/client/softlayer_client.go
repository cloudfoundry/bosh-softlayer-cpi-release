package client

import (
	"time"

	"github.com/softlayer/softlayer-go/session"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	SoftlayerAPIEndpointPublicDefault  = "https://api.softlayer.com/rest/v3.1"
	SoftlayerAPIEndpointPrivateDefault = "https://api.service.softlayer.com/rest/v3.1"
)

func NewSoftlayerClientSession(apiEndpoint string, username string, password string, trace bool, timeout int, logger boshlog.Logger) *session.Session {
	session := session.New(username, password, apiEndpoint)
	session.Debug = trace
	session.Timeout = time.Duration(timeout) * time.Second
	session.Logger = logger
	return session
}
