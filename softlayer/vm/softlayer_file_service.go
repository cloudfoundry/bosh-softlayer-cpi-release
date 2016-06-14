package vm

import (
	"bytes"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type SoftlayerFileService interface {
	SetVM(VM)
	Upload(string, []byte) error
	Download(string) ([]byte, error)
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

func (s *softlayerFileService) SetVM(vm VM) {
	s.vm = vm
}

func (s *softlayerFileService) Download(sourcePath string) ([]byte, error) {
	s.logger.Debug(s.logTag, "Downloading file at %s", sourcePath)

	buf := &bytes.Buffer{}
	err := s.sshClient.Download(ROOT_USER_NAME, s.vm.GetRootPassword(), s.vm.GetPrimaryBackendIP(), sourcePath, buf)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Download of %q failed", sourcePath)
	}

	s.logger.Debug(s.logTag, "Downloaded %d bytes", buf.Len())

	return buf.Bytes(), nil
}

func (s *softlayerFileService) Upload(destinationPath string, contents []byte) error {
	s.logger.Debug(s.logTag, "Uploading file to %s", destinationPath)

	buf := bytes.NewBuffer(contents)
	err := s.sshClient.Upload(ROOT_USER_NAME, s.vm.GetRootPassword(), s.vm.GetPrimaryBackendIP(), buf, destinationPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Upload to %q failed", destinationPath)
	}

	s.logger.Debug(s.logTag, "Upload complete")

	return nil
}
