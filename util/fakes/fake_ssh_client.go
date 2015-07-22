package util_fakes

type FakeSshClient struct {
	ExecCommandResult string
	ExecCommandError  error

	Username string
	Password string
	Ip       string
	Command  string
}

func (f *FakeSshClient) ExecCommand(username string, password string, ip string, command string) (string, error) {
	f.Username = username
	f.Password = password
	f.Ip = ip
	f.Command = command

	return f.ExecCommandResult, f.ExecCommandError
}

func GetFakeSshClient(fakeResult string, fakeError error) *FakeSshClient {
	return &FakeSshClient{
		ExecCommandResult: fakeResult,
		ExecCommandError:  fakeError,
	}
}
