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
	ExecCommand(command string) (string, error)

	Download(srcFile string, destination io.Writer) error
	DownloadFile(srcFile string, destFile string) error

	Upload(source io.Reader, destFile string) error
	UploadFile(srcFile string, destFile string) error
}

type sshClientImpl struct {
	ip       string
	username string
	password string
}

func (c *sshClientImpl) ExecCommand(command string) (string, error) {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", address(c.ip), config)
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

func (c *sshClientImpl) UploadFile(srcFile string, destFile string) error {
	source, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer source.Close()

	return c.Upload(source, destFile)
}

func (c *sshClientImpl) Upload(source io.Reader, destFile string) error {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", address(c.ip), config)
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

func (c *sshClientImpl) DownloadFile(srcFile string, destFile string) error {
	writer, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer writer.Close()

	return c.Download(srcFile, writer)
}

func (c *sshClientImpl) Download(srcFile string, destination io.Writer) error {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", address(c.ip), config)
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

func GetSshClient(username string, password string, ip string) SshClient {
	return &sshClientImpl{
		username: username,
		password: password,
		ip:       ip,
	}
}
