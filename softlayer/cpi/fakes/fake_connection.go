package cpi_fakes

import (
	"io"

	bslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi"
)

type FakeConnection struct {
	streamOutReturns    *io.Reader
	streamOutError      error
	streamOutCallCount  int
	streamOutArgsHandle string
	streamsOutArgsPath  string

	streamInArgsHandle string
	streamsInArgsPath  string
	streamInReader     io.Reader
	streamInCallCount  int

	createReturns bslcpi.Container

	runReturnsProcess bslcpi.Process
	runError          error
	runCallCount      int
	runHandle         string
	runProcessSpec    bslcpi.ProcessSpec
	runProcessIO      bslcpi.ProcessIO

	fakeError error

	createCallCount         int
	createCallContainerSpec bslcpi.ContainerSpec

	stopCallCount  int
	stopCallHandle string
	stopCallForce  bool
	stopCallError  error

	listCallReturns []string
	listCallCount   int
	listCallError   error

	destroyCallReturns string
	destroyCallCount   int
	destroyCallError   error
}

func (fc *FakeConnection) CreateReturns(vmId string, err error) {
	fc.fakeError = err
}

func (fc *FakeConnection) RunReturns(process bslcpi.Process, err error) {
	fc.runReturnsProcess = process
	fc.runError = err
	fc.runCallCount += 1
}

func (fc *FakeConnection) RunArgsForCall(arg int) (string, bslcpi.ProcessSpec, bslcpi.ProcessIO) {
	return fc.runHandle, fc.runProcessSpec, fc.runProcessIO
}

func (fc *FakeConnection) RunCallCount() int {
	return fc.runCallCount
}

func (fc *FakeConnection) StreamOutReturns(reader io.ReadCloser, err error) {
	//fc.streamOutReturns = reader
	fc.streamOutError = err
	fc.streamOutCallCount += 1
}

func (fc *FakeConnection) StreamOutArgsForCall(arg int) (string, string) {
	return fc.streamOutArgsHandle, fc.streamsOutArgsPath
}

func (fc *FakeConnection) StreamOutCallCount() int {
	return fc.streamOutCallCount
}

func (fc *FakeConnection) StreamInReturns(err error) {
	fc.streamOutError = err
	fc.streamOutCallCount += 1
}

func (fc *FakeConnection) StreamInArgsForCall(arg int) (string, string, io.Reader) {
	return fc.streamInArgsHandle, fc.streamsInArgsPath, fc.streamInReader
}

func (fc *FakeConnection) StreamInCallCount() int {
	return fc.streamInCallCount
}

func (fc *FakeConnection) CreateCallCount() int {
	return fc.createCallCount
}

func (fc *FakeConnection) CreateArgsForCall(arg int) bslcpi.ContainerSpec {
	return fc.createCallContainerSpec
}

func (fc *FakeConnection) StopReturns(err error) {
	fc.stopCallError = err
	fc.stopCallCount += 1
}

func (fc *FakeConnection) StopArgsForCall(arg int) (string, bool) {
	return fc.stopCallHandle, fc.stopCallForce
}

func (fc *FakeConnection) StopCallCount() int {
	return fc.stopCallCount
}

func (fc *FakeConnection) ListReturns(list []string, err error) {
	fc.listCallReturns = list
	fc.listCallError = err
	fc.listCallCount += 1
}

func (fc *FakeConnection) ListArgsForCall(arg int) []string {
	return fc.listCallReturns
}

func (fc *FakeConnection) ListCallCount() int {
	return fc.listCallCount
}

func (fc *FakeConnection) DestroyReturns(err error) {
	fc.destroyCallError = err
	fc.destroyCallCount += 1
}

func (fc *FakeConnection) DestroyArgsForCall(arg int) string {
	return fc.destroyCallReturns
}

func (fc *FakeConnection) DestroyCallCount() int {
	return fc.destroyCallCount
}
