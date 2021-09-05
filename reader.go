package s3iot

import (
	"io"
	"sync"
	"time"
)

type WaitReadInterceptorFactory struct {
	mu          sync.RWMutex
	waitPerByte time.Duration
}

func (f *WaitReadInterceptorFactory) WaitPerByte(w time.Duration) {
	f.mu.Lock()
	f.waitPerByte = w
	f.mu.Unlock()
}

type throttleReadInterceptor struct {
	factory *WaitReadInterceptorFactory
}

func (f *WaitReadInterceptorFactory) New() ReadInterceptor {
	return &throttleReadInterceptor{
		factory: f,
	}
}

func (i *throttleReadInterceptor) Reader(r io.ReadSeeker) io.ReadSeeker {
	return &throttleReader{
		ReadSeeker: r,
		factory:    i.factory,
	}
}

type throttleReader struct {
	io.ReadSeeker

	factory *WaitReadInterceptorFactory
}

func (r *throttleReader) Read(b []byte) (int, error) {
	r.factory.mu.RLock()
	waitPerByte := r.factory.waitPerByte
	r.factory.mu.RUnlock()
	time.Sleep(waitPerByte * time.Duration(len(b)))
	return r.ReadSeeker.Read(b)
}
