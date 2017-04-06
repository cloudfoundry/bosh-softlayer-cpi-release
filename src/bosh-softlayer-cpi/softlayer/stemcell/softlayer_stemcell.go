package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	sl "github.com/maximilien/softlayer-go/softlayer"
)

type SoftLayerStemcell struct {
	id   int
	uuid string

	softLayerFinder SoftLayerStemcellFinder
}

func NewSoftLayerStemcell(id int, uuid string, softLayerClient sl.Client, logger boshlog.Logger) SoftLayerStemcell {
	softLayerFinder := SoftLayerStemcellFinder{
		client: softLayerClient,
		logger: logger,
	}

	return SoftLayerStemcell{
		id:              id,
		uuid:            uuid,
		softLayerFinder: softLayerFinder,
	}
}

func (s SoftLayerStemcell) ID() int { return s.id }

func (s SoftLayerStemcell) Uuid() string { return s.uuid }

func (s SoftLayerStemcell) Delete() error {
	return nil
}
