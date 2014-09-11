package cpi

type Process interface {
	Wait() (int, error)
}
