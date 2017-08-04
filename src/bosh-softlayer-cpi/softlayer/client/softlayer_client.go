package client

import (
	"time"

	"github.com/softlayer/softlayer-go/session"
	"io"
	"log"
)

const (
	SoftlayerAPIEndpointPublicDefault  = "https://api.softlayer.com/rest/v3.1"
	SoftlayerAPIEndpointPrivateDefault = "https://api.service.softlayer.com/rest/v3.1"
	SoftlayerGoLogTag                  = "softlayerGo"
)

func NewSoftlayerClientSession(apiEndpoint string, username string, password string, trace bool, timeout int, writer io.Writer) *session.Session {
	session.Logger = log.New(writer, SoftlayerGoLogTag, log.LstdFlags)
	session := session.New(username, password, apiEndpoint)
	session.Debug = trace
	session.Timeout = time.Duration(timeout) * time.Second
	return session
}
