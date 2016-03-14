package test_helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	fakesshclient "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"
	fakesslclient "github.com/maximilien/softlayer-go/client/fakes"
)

func ReadJsonTestFixtures(workingDir, packageName, fileName string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(workingDir, "test_fixtures", packageName, fileName))
}

func SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient *fakesslclient.FakeSoftLayerClient, fileName string) {
	workingDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestResponse, err = ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", fileName)
	Expect(err).ToNot(HaveOccurred())
}

func SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient *fakesslclient.FakeSoftLayerClient, fileNames []string) {
	workingDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	for _, fileName := range fileNames {
		fileContents, err := ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", fileName)
		Expect(err).ToNot(HaveOccurred())

		fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestResponses = append(fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestResponses, fileContents)
	}
}

func SetTestFixturesForFakeSSHClient(fakeSSHClient *fakesshclient.FakeSshClient, results []string, err error) {
	for _, result := range results {
		fakeSSHClient.ExecCommandResults = append(fakeSSHClient.ExecCommandResults, result)
		fakeSSHClient.ExecCommandError = err
	}
}
