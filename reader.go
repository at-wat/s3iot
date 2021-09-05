package s3iot

import (
	"io"
	"sync"
	"time"
)

type WaitReadInterceptorFactory struct {
	mu           sync.RWMutex
	waitPerByte  time.Duration
	maxChunkSize int
}

func (f *WaitReadInterceptorFactory) SetWaitPerByte(w time.Duration) {
	f.mu.Lock()
	f.waitPerByte = w
	f.mu.Unlock()
}

func (f *WaitReadInterceptorFactory) SetMaxChunkSize(s int) {
	f.mu.Lock()
	f.maxChunkSize = s
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
	maxChunkSize := r.factory.maxChunkSize
	r.factory.mu.RUnlock()

	if maxChunkSize != 0 && len(b) > maxChunkSize {
		b = b[:maxChunkSize]
	}

	n, err := r.ReadSeeker.Read(b)
	time.Sleep(waitPerByte * time.Duration(n))
	return n, err
}
