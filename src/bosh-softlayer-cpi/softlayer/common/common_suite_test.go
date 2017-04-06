package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Suite")
}

type FailingWriteCloser struct {
	WriteErr error
	CloseErr error
}

func (wc FailingWriteCloser) Write(data []byte) (int, error) { return len(data), wc.WriteErr }
func (wc FailingWriteCloser) Close() error                   { return wc.CloseErr }
