package cpi

import (
	"io"
	"os"
)

type Container interface {
	Handle() string
	Run(processSpec ProcessSpec, processIO ProcessIO) (Process, error)
	Stop(bool) error

	StreamOut(path string) (*os.File, error)
	StreamIn(path string, reader io.Reader) error
}
