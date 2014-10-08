package test_helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	fakesslclient "github.com/maximilien/softlayer-go/client/fakes"
)

func ReadJsonTestFixtures(workingDir, packageName, fileName string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(workingDir, "test_fixtures", packageName, fileName))
}

func SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient *fakesslclient.FakeSoftLayerClient, fileName string) {
	workingDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	fakeSoftLayerClient.DoRawHttpRequestResponse, err = ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", fileName)
	Expect(err).ToNot(HaveOccurred())
}
