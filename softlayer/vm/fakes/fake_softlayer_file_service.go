package fakes

type FakeSoftlayerFileService struct {
	UploadInputs []UploadInput
	UploadErr    error

	DownloadSourcePath string
	DownloadContents   []byte
	DownloadErr        error
}

type UploadInput struct {
	DestinationPath string
	Contents        []byte
}

func NewFakeSoftlayerFileService() *FakeSoftlayerFileService {
	return &FakeSoftlayerFileService{
		UploadInputs: []UploadInput{},
	}
}

func (s *FakeSoftlayerFileService) Upload(destinationPath string, contents []byte) error {
	s.UploadInputs = append(s.UploadInputs, UploadInput{
		DestinationPath: destinationPath,
		Contents:        contents,
	})

	return s.UploadErr
}

func (s *FakeSoftlayerFileService) Download(sourcePath string) ([]byte, error) {
	s.DownloadSourcePath = sourcePath

	return s.DownloadContents, s.DownloadErr
}
