package fakes

import (
	bslcaction "github.com/cloudfoundry/bosh-softlayer-cpi/action"
)

type FakeCaller struct {
	CallAction bslcaction.Action
	CallArgs   []interface{}
	CallResult interface{}
	CallErr    error
}

func (caller *FakeCaller) Call(action bslcaction.Action, args []interface{}) (interface{}, error) {
	caller.CallAction = action
	caller.CallArgs = args
	return caller.CallResult, caller.CallErr
}
