package client

import (
	"github.com/ncw/swift"
	"time"
)

func NewSwiftClient(swiftEndpoint string, swiftUsername string, swiftPassword string, timeoutSec int, retries int) *swift.Connection {
	// Create a connection
	conn := &swift.Connection{
		UserName: swiftUsername,
		ApiKey:   swiftPassword,
		AuthUrl:  swiftEndpoint,
		Timeout:  time.Duration(timeoutSec) * time.Second, // Connection timeout is 10s
		Retries:  retries,
	}

	// If you don't call c.Authenticate() before calling one of the connection methods then it will be called for you on the first access.
	return conn
}
