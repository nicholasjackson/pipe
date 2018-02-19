// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package providers

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/hashicorp/go-hclog"
	"sync"
)

var (
	lockProviderMockListen sync.RWMutex
	lockProviderMockSetup  sync.RWMutex
	lockProviderMockStop   sync.RWMutex
	lockProviderMockType   sync.RWMutex
)

// ProviderMock is a mock implementation of Provider.
//
//     func TestSomethingThatUsesProvider(t *testing.T) {
//
//         // make and configure a mocked Provider
//         mockedProvider := &ProviderMock{
//             ListenFunc: func() (<-chan *Message, error) {
// 	               panic("TODO: mock out the Listen method")
//             },
//             SetupFunc: func(cp ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
// 	               panic("TODO: mock out the Setup method")
//             },
//             StopFunc: func() error {
// 	               panic("TODO: mock out the Stop method")
//             },
//             TypeFunc: func() string {
// 	               panic("TODO: mock out the Type method")
//             },
//         }
//
//         // TODO: use mockedProvider in code that requires Provider
//         //       and then make assertions.
//
//     }
type ProviderMock struct {
	// ListenFunc mocks the Listen method.
	ListenFunc func() (<-chan *Message, error)

	// SetupFunc mocks the Setup method.
	SetupFunc func(cp ConnectionPool, log hclog.Logger, stats *statsd.Client) error

	// StopFunc mocks the Stop method.
	StopFunc func() error

	// TypeFunc mocks the Type method.
	TypeFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Listen holds details about calls to the Listen method.
		Listen []struct {
		}
		// Setup holds details about calls to the Setup method.
		Setup []struct {
			// Cp is the cp argument value.
			Cp ConnectionPool
			// Log is the log argument value.
			Log hclog.Logger
			// Stats is the stats argument value.
			Stats *statsd.Client
		}
		// Stop holds details about calls to the Stop method.
		Stop []struct {
		}
		// Type holds details about calls to the Type method.
		Type []struct {
		}
	}
}

// Listen calls ListenFunc.
func (mock *ProviderMock) Listen() (<-chan *Message, error) {
	if mock.ListenFunc == nil {
		panic("moq: ProviderMock.ListenFunc is nil but Provider.Listen was just called")
	}
	callInfo := struct {
	}{}
	lockProviderMockListen.Lock()
	mock.calls.Listen = append(mock.calls.Listen, callInfo)
	lockProviderMockListen.Unlock()
	return mock.ListenFunc()
}

// ListenCalls gets all the calls that were made to Listen.
// Check the length with:
//     len(mockedProvider.ListenCalls())
func (mock *ProviderMock) ListenCalls() []struct {
} {
	var calls []struct {
	}
	lockProviderMockListen.RLock()
	calls = mock.calls.Listen
	lockProviderMockListen.RUnlock()
	return calls
}

// Setup calls SetupFunc.
func (mock *ProviderMock) Setup(cp ConnectionPool, log hclog.Logger, stats *statsd.Client) error {
	if mock.SetupFunc == nil {
		panic("moq: ProviderMock.SetupFunc is nil but Provider.Setup was just called")
	}
	callInfo := struct {
		Cp    ConnectionPool
		Log   hclog.Logger
		Stats *statsd.Client
	}{
		Cp:    cp,
		Log:   log,
		Stats: stats,
	}
	lockProviderMockSetup.Lock()
	mock.calls.Setup = append(mock.calls.Setup, callInfo)
	lockProviderMockSetup.Unlock()
	return mock.SetupFunc(cp, log, stats)
}

// SetupCalls gets all the calls that were made to Setup.
// Check the length with:
//     len(mockedProvider.SetupCalls())
func (mock *ProviderMock) SetupCalls() []struct {
	Cp    ConnectionPool
	Log   hclog.Logger
	Stats *statsd.Client
} {
	var calls []struct {
		Cp    ConnectionPool
		Log   hclog.Logger
		Stats *statsd.Client
	}
	lockProviderMockSetup.RLock()
	calls = mock.calls.Setup
	lockProviderMockSetup.RUnlock()
	return calls
}

// Stop calls StopFunc.
func (mock *ProviderMock) Stop() error {
	if mock.StopFunc == nil {
		panic("moq: ProviderMock.StopFunc is nil but Provider.Stop was just called")
	}
	callInfo := struct {
	}{}
	lockProviderMockStop.Lock()
	mock.calls.Stop = append(mock.calls.Stop, callInfo)
	lockProviderMockStop.Unlock()
	return mock.StopFunc()
}

// StopCalls gets all the calls that were made to Stop.
// Check the length with:
//     len(mockedProvider.StopCalls())
func (mock *ProviderMock) StopCalls() []struct {
} {
	var calls []struct {
	}
	lockProviderMockStop.RLock()
	calls = mock.calls.Stop
	lockProviderMockStop.RUnlock()
	return calls
}

// Type calls TypeFunc.
func (mock *ProviderMock) Type() string {
	if mock.TypeFunc == nil {
		panic("moq: ProviderMock.TypeFunc is nil but Provider.Type was just called")
	}
	callInfo := struct {
	}{}
	lockProviderMockType.Lock()
	mock.calls.Type = append(mock.calls.Type, callInfo)
	lockProviderMockType.Unlock()
	return mock.TypeFunc()
}

// TypeCalls gets all the calls that were made to Type.
// Check the length with:
//     len(mockedProvider.TypeCalls())
func (mock *ProviderMock) TypeCalls() []struct {
} {
	var calls []struct {
	}
	lockProviderMockType.RLock()
	calls = mock.calls.Type
	lockProviderMockType.RUnlock()
	return calls
}