package util_fakes

type FakeSshClient struct {
	ExecCommandResult string
	ExecCommandError  error
	UploadFileError   error
	DownloadFileError error
}

func (f *FakeSshClient) ExecCommand(username string, password string, ip string, command string) (string, error) {
	return f.ExecCommandResult, f.ExecCommandError
}
func (f *FakeSshClient) UploadFile(username string, password string, ip string, srcFile string, destFile string) error {
	return f.UploadFileError
}
func (f *FakeSshClient) DownloadFile(username string, password string, ip string, srcFile string, destFile string) error {
	return f.DownloadFileError
}

func GetFakeSshClient(fakeResult string, fakeError error) *FakeSshClient {
	return &FakeSshClient{
		ExecCommandResult: fakeResult,
		ExecCommandError:  fakeError,
	}
}
