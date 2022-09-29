package fakes

import "sync"

type DeprecationDate struct {
	GetDateCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Feed    string
			Version string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string, string) (string, error)
	}
}

func (f *DeprecationDate) GetDate(param1 string, param2 string) (string, error) {
	f.GetDateCall.mutex.Lock()
	defer f.GetDateCall.mutex.Unlock()
	f.GetDateCall.CallCount++
	f.GetDateCall.Receives.Feed = param1
	f.GetDateCall.Receives.Version = param2
	if f.GetDateCall.Stub != nil {
		return f.GetDateCall.Stub(param1, param2)
	}
	return f.GetDateCall.Returns.String, f.GetDateCall.Returns.Error
}
