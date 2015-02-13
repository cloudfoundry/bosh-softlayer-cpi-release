package util

import (
	"code.google.com/p/go.crypto/ssh"
)

type SshClient interface {
	ExecCommand(username string, password string, ip string, command string) (string, error)
}

type sshClientImpl struct{}

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

func GetSshClient() SshClient {
	return &sshClientImpl{}
}
