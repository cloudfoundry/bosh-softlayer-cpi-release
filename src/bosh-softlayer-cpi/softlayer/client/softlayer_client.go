package client

import (
	"log"
	"time"

	"github.com/softlayer/softlayer-go/session"
)

const (
	SoftlayerAPIEndpointPublicDefault  = "https://api.softlayer.com/rest/v3.1"
	SoftlayerAPIEndpointPrivateDefault = "https://api.service.softlayer.com/rest/v3.1"
	SoftlayerGoLogTag                  = "[softlayerGo] "
)

func NewSoftlayerClientSession(apiEndpoint string, username string, password string, trace bool, timeoutSec int, retries int, retryWaitSec int, outLogger *log.Logger) *session.Session {
	// Use native logger's prefix as bosh logger tag
	outLogger.SetPrefix(outLogger.Prefix() + SoftlayerGoLogTag)
	session.Logger = outLogger
	session := session.New(username, password, apiEndpoint)
	session.Debug = trace
	session.Timeout = time.Duration(timeoutSec) * time.Second
	session.Retries = retries
	session.RetryWait = time.Duration(retryWaitSec) * time.Second
	return session
}
