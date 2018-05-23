package api

import (
	"bytes"

	"bosh-softlayer-cpi/logger"
)

type MultiLogger struct {
	logger.Logger
	LogBuff *bytes.Buffer
}
