package vps_smoke_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestVpsSmoke(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VpsSmoke Suite")
}

var _ = BeforeSuite(func() {
	vps := "test_assets/vps"
	postgresURL, err := GetPostgresURL()
	Expect(err).ToNot(HaveOccurred())

	command := exec.Command(vps, "--scheme=https", "--tls-host=0.0.0.0", "--tls-port=1443", "--tls-certificate=test_assets/server.pem", "--tls-key=test_assets/server.key", "--databaseDriver", "postgres", "--databaseConnectionString", postgresURL)
	_, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.KillAndWait()
})

func GetPostgresURL() (string, error) {
	TargetURL := os.Getenv("POSTGRES_URL")
	if TargetURL == "" {
		return "", errors.New("POSTGRES_URL environment must be set")
	}

	return TargetURL, nil
}
