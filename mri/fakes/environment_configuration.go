package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit"
)

type EnvironmentConfiguration struct {
	ConfigureCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Env  packit.Environment
			Path string
		}
		Returns struct {
			Error error
		}
		Stub func(packit.Environment, string) error
	}
}

func (f *EnvironmentConfiguration) Configure(param1 packit.Environment, param2 string) error {
	f.ConfigureCall.Lock()
	defer f.ConfigureCall.Unlock()
	f.ConfigureCall.CallCount++
	f.ConfigureCall.Receives.Env = param1
	f.ConfigureCall.Receives.Path = param2
	if f.ConfigureCall.Stub != nil {
		return f.ConfigureCall.Stub(param1, param2)
	}
	return f.ConfigureCall.Returns.Error
}
