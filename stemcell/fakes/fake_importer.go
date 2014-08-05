package fakes

import (
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/stemcell"
)

type FakeImporter struct {
	ImportFromPathImagePath string
	ImportFromPathStemcell  bslcstem.Stemcell
	ImportFromPathErr       error
}

func (c *FakeImporter) ImportFromPath(imagePath string) (bslcstem.Stemcell, error) {
	c.ImportFromPathImagePath = imagePath
	return c.ImportFromPathStemcell, c.ImportFromPathErr
}
