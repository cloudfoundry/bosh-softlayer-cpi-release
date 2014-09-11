package cpi

var (
	BindMountModeRW     int
	BindMountOriginHost string
)

type ProcessSpec struct {
	Path       string
	Args       []string
	Privileged bool
}

type Properties struct {
}

type BindMount struct {
	SrcPath string
	DstPath string
	Mode    int
	Origin  string
}

type ContainerSpec struct {
	Handle     string
	RootFSPath string
	Network    string
	BindMounts []BindMount
	Properties Properties
}
