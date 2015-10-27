package concurrency_sqlite_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSqliteConcurrency(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration: concurrency_sqlite Suite")
}
