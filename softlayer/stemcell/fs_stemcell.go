package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const fsStemcellLogTag = "FSStemcell"

type FSStemcell struct {
	id      string
	logger boshlog.Logger
}

func NewFSStemcell(id string, logger boshlog.Logger) FSStemcell {
	return FSStemcell{id: id, logger: logger}
}

func (s FSStemcell) ID() string { return s.id }

func (s FSStemcell) Delete() error {
	return nil
}
