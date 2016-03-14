package util

import (
	"errors"
	"io/ioutil"
	"regexp"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

type SshClient interface {
	ExecCommand(username string, password string, ip string, command string) (string, error)
	UploadFile(username string, password string, ip string, srcFile string, destFile string) error
	DownloadFile(username string, password string, ip string, srcFile string, destFile string) error
}

type sshClientImpl struct{}

func IsIP(ip string) (b bool) {
	if m, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", ip); !m {
		return false
	}
	return true
}

func IsDir(d string) (b bool) {
	if m, _ := regexp.MatchString("^/.*/$", d); !m {
		return false
	}
	return true
}

func (c *sshClientImpl) ExecCommand(username string, password string, ip string, command string) (string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	client, err := ssh.Dial("tcp", ip+":22", config)

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	output, err := session.Output(command)

	return string(output), err
}

func (c *sshClientImpl) UploadFile(username string, password string, ip string, srcFile string, destFile string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	if !IsIP(ip) {
		return errors.New("invalid IP address")
	}

	if IsDir(srcFile) || IsDir(destFile) {
		return errors.New("Is a directory")
	}

	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	data, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}

	f, err := sftp.Create(destFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(data))
	if err != nil {
		return err
	}

	_, err = sftp.Lstat(destFile)
	if err != nil {
		return err
	}
	return nil
}

func (c *sshClientImpl) DownloadFile(username string, password string, ip string, srcFile string, destFile string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	if !IsIP(ip) {
		return errors.New("invalid IP address")
	}

	if IsDir(srcFile) || IsDir(destFile) {
		return errors.New("Is a directory")
	}

	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	pFile, err := sftp.Open(srcFile)
	if err != nil {
		return err
	}
	defer pFile.Close()

	data, err := ioutil.ReadAll(pFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destFile, data, 0755)
	if err != nil {
		return err
	}

	return nil
}

func GetSshClient() SshClient {
	return &sshClientImpl{}
}
