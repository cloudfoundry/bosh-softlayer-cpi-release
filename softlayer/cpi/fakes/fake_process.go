package cpi_fakes

type FakeProcess struct {
	waitReturns int
	waitError   error
}

func (fp *FakeProcess) Wait() (int, error) {
	return fp.waitReturns, fp.waitError
}

func (fp *FakeProcess) WaitReturns(returnCode int, returnError error) {
	fp.waitReturns = returnCode
	fp.waitError = returnError
}
