package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SoftLayerStemcell struct {
	id   int
	uuid string
}

func NewSoftLayerStemcell(id int, uuid string, logger boshlog.Logger) SoftLayerStemcell {
	return SoftLayerStemcell{
		id:   id,
		uuid: uuid,
	}
}

func (ss SoftLayerStemcell) ID() int { return ss.id }

func (ss SoftLayerStemcell) Uuid() string { return ss.uuid }
