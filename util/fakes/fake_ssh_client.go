package util_fakes

type FakeSshClient struct {
	ExecCommandResultsIndex int
	ExecCommandResults      []string
	ExecCommandError        error

	Username string
	Password string
	Ip       string
	Command  string

	UploadFileError   error
	DownloadFileError error
}

func NewFakeSshClient() *FakeSshClient {
	fssh := &FakeSshClient{
		ExecCommandResults: []string{},
		ExecCommandError:   nil,
	}
	return fssh
}

func (f *FakeSshClient) ExecCommand(username string, password string, ip string, command string) (string, error) {
	f.Username = username
	f.Password = password
	f.Ip = ip
	f.Command = command
	f.ExecCommandResultsIndex = f.ExecCommandResultsIndex + 1
	return f.ExecCommandResults[f.ExecCommandResultsIndex-1], f.ExecCommandError
}

func (f *FakeSshClient) UploadFile(username string, password string, ip string, srcFile string, destFile string) error {
	f.Username = username
	f.Password = password
	f.Ip = ip
	return f.UploadFileError
}

func (f *FakeSshClient) DownloadFile(username string, password string, ip string, srcFile string, destFile string) error {
	f.Username = username
	f.Password = password
	f.Ip = ip
	return f.DownloadFileError
}

func GetFakeSshClient(fakeResults []string, fakeError error) *FakeSshClient {
	return &FakeSshClient{
		ExecCommandResults: fakeResults,
		ExecCommandError:   fakeError,
	}
}
