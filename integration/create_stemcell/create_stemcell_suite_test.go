package create_stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCreateStemcell(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: create_stemcell Suite")
}
