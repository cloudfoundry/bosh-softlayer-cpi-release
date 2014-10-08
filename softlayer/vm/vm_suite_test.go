package vm_test

import (
	"os"
	"path/filepath"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	common "github.com/maximilien/bosh-softlayer-cpi/common"
)

func TestVM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vm Suite")
}

type NonJSONMarshable struct{}

func (m NonJSONMarshable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("fake-marshal-err")
}

type FailingWriteCloser struct {
	WriteErr error
	CloseErr error
}

func (wc FailingWriteCloser) Write(data []byte) (int, error) { return len(data), wc.WriteErr }
func (wc FailingWriteCloser) Close() error                   { return wc.CloseErr }

func SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient, fileName string) {
	workingDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	fakeSoftLayerClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures(filepath.Join(workingDir, "..", ".."), "softlayer", fileName)
	Expect(err).ToNot(HaveOccurred())
}
