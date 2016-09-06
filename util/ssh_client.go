package util

import (
	"io"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//go:generate counterfeiter -o fakes/fake_ssh_client.go . SshClient
type SshClient interface {
	ExecCommand(username string, password string, ip string, command string) (string, error)

	Download(username, password, ip, srcFile string, destination io.Writer) error
	DownloadFile(username string, password string, ip string, srcFile string, destFile string) error

	Upload(username, password, ip string, source io.Reader, destFile string) error
	UploadFile(username string, password string, ip string, srcFile string, destFile string) error
}

type sshClientImpl struct{}

func (c *sshClientImpl) ExecCommand(username string, password string, ip string, command string) (string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	client, err := ssh.Dial("tcp", address(ip), config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.Output(command)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (c *sshClientImpl) UploadFile(username string, password string, ip string, srcFile string, destFile string) error {
	source, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer source.Close()

	return c.Upload(username, password, ip, source, destFile)
}

func (c *sshClientImpl) Upload(username, password, ip string, source io.Reader, destFile string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	client, err := ssh.Dial("tcp", address(ip), config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	f, err := sftp.Create(destFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.ReadFrom(source)
	if err != nil {
		return err
	}

	return nil
}

func (c *sshClientImpl) DownloadFile(username string, password string, ip string, srcFile string, destFile string) error {
	writer, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer writer.Close()

	return c.Download(username, password, ip, srcFile, writer)
}

func (c *sshClientImpl) Download(username, password, ip, srcFile string, destination io.Writer) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	client, err := ssh.Dial("tcp", address(ip), config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	f, err := sftp.Open(srcFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteTo(destination)
	if err != nil {
		return err
	}

	return nil
}

func address(a string) string {
	if _, _, err := net.SplitHostPort(a); err != nil {
		a = net.JoinHostPort(a, "22")
	}

	return a
}

func GetSshClient() SshClient {
	return &sshClientImpl{}
}
