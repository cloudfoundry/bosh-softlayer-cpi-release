package api

import (
	"time"
)

var (
	TIMEOUT                   time.Duration = 240 * time.Minute
	POLLING_INTERVAL          time.Duration = 5 * time.Second
	LocalDiskFlagNotSet       bool
	LocalDNSConfigurationFile string
	NetworkInterface          string
	LengthOfHostName          int
)
