package vm

type NotSupportedError struct{}

func (e NotSupportedError) Type() string  { return "Bosh::Clouds::NotSupported" }
func (e NotSupportedError) Error() string { return "Not supported" }
