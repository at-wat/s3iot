package s3iot

import (
	"io"
	"sync"
	"time"
)

const (
	DefaultWaitReadInterceptorMaxChunkSize = 4 * 1024
)

type WaitReadInterceptorFactory struct {
	mu           sync.RWMutex
	waitPerByte  time.Duration
	maxChunkSize int
}

func NewWaitReadInterceptorFactory(waitPerByte time.Duration) *WaitReadInterceptorFactory {
	return &WaitReadInterceptorFactory{
		waitPerByte:  waitPerByte,
		maxChunkSize: DefaultWaitReadInterceptorMaxChunkSize,
	}
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

type waitReadInterceptor struct {
	factory *WaitReadInterceptorFactory
}

func (f *WaitReadInterceptorFactory) New() ReadInterceptor {
	return &waitReadInterceptor{
		factory: f,
	}
}

func (i *waitReadInterceptor) Reader(r io.ReadSeeker) io.ReadSeeker {
	return &waitReader{
		ReadSeeker: r,
		factory:    i.factory,
	}
}

type waitReader struct {
	io.ReadSeeker

	factory *WaitReadInterceptorFactory
}

func (r *waitReader) Read(b []byte) (int, error) {
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
