package delete_stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDeleteStemcell(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: delete_stemcell Suite")
}
