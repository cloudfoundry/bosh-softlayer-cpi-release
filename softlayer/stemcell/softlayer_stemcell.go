package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

const softLayerStemcellLogTag = "SoftLayerStemcell"

type SoftLayerStemcell struct {
	id     int
	uuid   string
	logger boshlog.Logger
}

func NewSoftLayerStemcell(id int, uuid string, logger boshlog.Logger) SoftLayerStemcell {
	return SoftLayerStemcell{id: id, logger: logger}
}

func (s SoftLayerStemcell) ID() int { return s.id }

func (s SoftLayerStemcell) Uuid() string { return s.uuid }

func (s SoftLayerStemcell) Delete() error {
	//TODO: implement me
	return nil
}
