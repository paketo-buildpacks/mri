package fakes

import "sync"

type License struct {
	LookupLicensesCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			DependencyName string
			SourceURL      string
		}
		Returns struct {
			InterfaceSlice []interface {
			}
			Error error
		}
		Stub func(string, string) ([]interface {
		}, error)
	}
}

func (f *License) LookupLicenses(param1 string, param2 string) ([]interface {
}, error) {
	f.LookupLicensesCall.mutex.Lock()
	defer f.LookupLicensesCall.mutex.Unlock()
	f.LookupLicensesCall.CallCount++
	f.LookupLicensesCall.Receives.DependencyName = param1
	f.LookupLicensesCall.Receives.SourceURL = param2
	if f.LookupLicensesCall.Stub != nil {
		return f.LookupLicensesCall.Stub(param1, param2)
	}
	return f.LookupLicensesCall.Returns.InterfaceSlice, f.LookupLicensesCall.Returns.Error
}
