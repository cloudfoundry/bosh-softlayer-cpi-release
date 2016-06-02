package vm

import (
	"bytes"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	datatypes "github.com/maximilien/softlayer-go/data_types"

	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
)

type SoftlayerFileService interface {
	Upload(string, []byte) error
	Download(string) ([]byte, error)
}

type softlayerFileService struct {
	sshClient    util.SshClient
	virtualGuest datatypes.SoftLayer_Virtual_Guest
	logger       boshlog.Logger
	logTag       string
}

func NewSoftlayerFileService(
	sshClient util.SshClient,
	virtualGuest datatypes.SoftLayer_Virtual_Guest,
	logger boshlog.Logger,
) SoftlayerFileService {
	return &softlayerFileService{
		sshClient:    sshClient,
		virtualGuest: virtualGuest,
		logger:       logger,
		logTag:       "softlayerFileService",
	}
}

func (s *softlayerFileService) Download(sourcePath string) ([]byte, error) {
	s.logger.Debug(s.logTag, "Downloading file at %s", sourcePath)

	buf := &bytes.Buffer{}
	err := s.sshClient.Download(ROOT_USER_NAME, s.getRootPassword(s.virtualGuest), s.virtualGuest.PrimaryBackendIpAddress, sourcePath, buf)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Download of %q failed", sourcePath)
	}

	s.logger.Debug(s.logTag, "Downloaded %d bytes", buf.Len())

	return buf.Bytes(), nil
}

func (s *softlayerFileService) Upload(destinationPath string, contents []byte) error {
	s.logger.Debug(s.logTag, "Uploading file to %s", destinationPath)

	buf := bytes.NewBuffer(contents)
	err := s.sshClient.Upload(ROOT_USER_NAME, s.getRootPassword(s.virtualGuest), s.virtualGuest.PrimaryBackendIpAddress, buf, destinationPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Upload to %q failed", destinationPath)
	}

	s.logger.Debug(s.logTag, "Upload complete")

	return nil
}

func (s *softlayerFileService) getRootPassword(virtualGuest datatypes.SoftLayer_Virtual_Guest) string {
	passwords := virtualGuest.OperatingSystem.Passwords
	for _, password := range passwords {
		if password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}

	return ""
}
