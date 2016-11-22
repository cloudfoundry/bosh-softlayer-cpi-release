package common

import (
	"bytes"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type SoftlayerFileService interface {
	Upload(user string, password string, target string, destinationPath string, contents []byte) error
	Download(user string, password string, target string, sourcePath string) ([]byte, error)
}

type softlayerFileService struct {
	sshClient util.SshClient
	vm        VM
	logger    boshlog.Logger
	logTag    string
}

func NewSoftlayerFileService(sshClient util.SshClient, logger boshlog.Logger) SoftlayerFileService {
	return &softlayerFileService{
		sshClient: sshClient,
		logger:    logger,
		logTag:    "softlayerFileService",
	}
}

func (s *softlayerFileService) Download(user string, password string, target string, sourcePath string) ([]byte, error) {
	s.logger.Debug(s.logTag, "Downloading file at %s", sourcePath)

	buf := &bytes.Buffer{}
	err := s.sshClient.Download(user, password, target, sourcePath, buf)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Download of %q failed", sourcePath)
	}

	s.logger.Debug(s.logTag, "Downloaded %d bytes", buf.Len())

	return buf.Bytes(), nil
}

func (s *softlayerFileService) Upload(user string, password string, target string, destinationPath string, contents []byte) error {
	s.logger.Debug(s.logTag, "Uploading file to %s", destinationPath)

	buf := bytes.NewBuffer(contents)
	err := s.sshClient.Upload(user, password, target, buf, destinationPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Upload to %q failed", destinationPath)
	}

	s.logger.Debug(s.logTag, "Upload complete")

	return nil
}
