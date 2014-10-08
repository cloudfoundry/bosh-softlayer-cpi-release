package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const softLayerStemcellLogTag = "FSStemcell"

type SoftLayerStemcell struct {
	id     string
	logger boshlog.Logger
}

func NewSoftLayerStemcell(id string, logger boshlog.Logger) SoftLayerStemcell {
	return SoftLayerStemcell{id: id, logger: logger}
}

func (s SoftLayerStemcell) ID() string { return s.id }

func (s SoftLayerStemcell) Delete() error {
	//TODO: implement me
	return nil
}
