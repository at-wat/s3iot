// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock_s3iot

import (
	"github.com/at-wat/s3iot"
	"io"
	"sync"
)

// Ensure, that MockReadInterceptorFactory does implement s3iot.ReadInterceptorFactory.
// If this is not the case, regenerate this file with moq.
var _ s3iot.ReadInterceptorFactory = &MockReadInterceptorFactory{}

// MockReadInterceptorFactory is a mock implementation of s3iot.ReadInterceptorFactory.
//
//	func TestSomethingThatUsesReadInterceptorFactory(t *testing.T) {
//
//		// make and configure a mocked s3iot.ReadInterceptorFactory
//		mockedReadInterceptorFactory := &MockReadInterceptorFactory{
//			NewFunc: func() s3iot.ReadInterceptor {
//				panic("mock out the New method")
//			},
//		}
//
//		// use mockedReadInterceptorFactory in code that requires s3iot.ReadInterceptorFactory
//		// and then make assertions.
//
//	}
type MockReadInterceptorFactory struct {
	// NewFunc mocks the New method.
	NewFunc func() s3iot.ReadInterceptor

	// calls tracks calls to the methods.
	calls struct {
		// New holds details about calls to the New method.
		New []struct {
		}
	}
	lockNew sync.RWMutex
}

// New calls NewFunc.
func (mock *MockReadInterceptorFactory) New() s3iot.ReadInterceptor {
	if mock.NewFunc == nil {
		panic("MockReadInterceptorFactory.NewFunc: method is nil but ReadInterceptorFactory.New was just called")
	}
	callInfo := struct {
	}{}
	mock.lockNew.Lock()
	mock.calls.New = append(mock.calls.New, callInfo)
	mock.lockNew.Unlock()
	return mock.NewFunc()
}

// NewCalls gets all the calls that were made to New.
// Check the length with:
//
//	len(mockedReadInterceptorFactory.NewCalls())
func (mock *MockReadInterceptorFactory) NewCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockNew.RLock()
	calls = mock.calls.New
	mock.lockNew.RUnlock()
	return calls
}

// Ensure, that MockReadInterceptor does implement s3iot.ReadInterceptor.
// If this is not the case, regenerate this file with moq.
var _ s3iot.ReadInterceptor = &MockReadInterceptor{}

// MockReadInterceptor is a mock implementation of s3iot.ReadInterceptor.
//
//	func TestSomethingThatUsesReadInterceptor(t *testing.T) {
//
//		// make and configure a mocked s3iot.ReadInterceptor
//		mockedReadInterceptor := &MockReadInterceptor{
//			ReaderFunc: func(readSeeker io.ReadSeeker) io.ReadSeeker {
//				panic("mock out the Reader method")
//			},
//		}
//
//		// use mockedReadInterceptor in code that requires s3iot.ReadInterceptor
//		// and then make assertions.
//
//	}
type MockReadInterceptor struct {
	// ReaderFunc mocks the Reader method.
	ReaderFunc func(readSeeker io.ReadSeeker) io.ReadSeeker

	// calls tracks calls to the methods.
	calls struct {
		// Reader holds details about calls to the Reader method.
		Reader []struct {
			// ReadSeeker is the readSeeker argument value.
			ReadSeeker io.ReadSeeker
		}
	}
	lockReader sync.RWMutex
}

// Reader calls ReaderFunc.
func (mock *MockReadInterceptor) Reader(readSeeker io.ReadSeeker) io.ReadSeeker {
	if mock.ReaderFunc == nil {
		panic("MockReadInterceptor.ReaderFunc: method is nil but ReadInterceptor.Reader was just called")
	}
	callInfo := struct {
		ReadSeeker io.ReadSeeker
	}{
		ReadSeeker: readSeeker,
	}
	mock.lockReader.Lock()
	mock.calls.Reader = append(mock.calls.Reader, callInfo)
	mock.lockReader.Unlock()
	return mock.ReaderFunc(readSeeker)
}

// ReaderCalls gets all the calls that were made to Reader.
// Check the length with:
//
//	len(mockedReadInterceptor.ReaderCalls())
func (mock *MockReadInterceptor) ReaderCalls() []struct {
	ReadSeeker io.ReadSeeker
} {
	var calls []struct {
		ReadSeeker io.ReadSeeker
	}
	mock.lockReader.RLock()
	calls = mock.calls.Reader
	mock.lockReader.RUnlock()
	return calls
}