package util_fakes

type FakeSshClient struct {
	ExecCommandResult string
	ExecCommandError  error
}

func (f *FakeSshClient) ExecCommand(username string, password string, ip string, command string) (string, error) {
	return f.ExecCommandResult, f.ExecCommandError
}

func GetFakeSshClient(fakeResult string, fakeError error) *FakeSshClient {
	return &FakeSshClient{
		ExecCommandResult: fakeResult,
		ExecCommandError:  fakeError,
	}
}
