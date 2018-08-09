package action

type InfoResult struct {
	StemcellFormats []string `json:"stemcell_formats"`
}

type InfoAction struct{}

func NewInfo() (action InfoAction) {
	return
}

func (a InfoAction) Run() (InfoResult, error) {
	return InfoResult{
		StemcellFormats: []string{
			"softlayer-legacy-light",
		},
	}, nil
}
