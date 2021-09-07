// Copyright 2021 The s3iot authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3iot

import (
	"io"
	"sync"
	"time"
)

// Default ReadInterceptor parameters.
const (
	DefaultWaitReadInterceptorMaxChunkSize = 4 * 1024
)

// WaitReadInterceptorFactory creates WaitInterceptor.
type WaitReadInterceptorFactory struct {
	mu           sync.RWMutex
	waitPerByte  time.Duration
	maxChunkSize int
}

// WaitReadInterceptorOption configures WaitReadInterceptorFactory.
type WaitReadInterceptorOption func(*WaitReadInterceptorFactory)

// WaitReadInterceptorMaxChunkSize sets MaxChunkSize.
func WaitReadInterceptorMaxChunkSize(s int) WaitReadInterceptorOption {
	return func(f *WaitReadInterceptorFactory) {
		f.mu.Lock()
		f.maxChunkSize = s
		f.mu.Unlock()
	}
}

// NewWaitReadInterceptorFactory creates WaitReadInterceptorFactory with wait/byte value.
func NewWaitReadInterceptorFactory(waitPerByte time.Duration, opts ...WaitReadInterceptorOption) *WaitReadInterceptorFactory {
	f := &WaitReadInterceptorFactory{
		waitPerByte:  waitPerByte,
		maxChunkSize: DefaultWaitReadInterceptorMaxChunkSize,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// SetWaitPerByte sets wait/byte parameter.
// It can be changed during upload.
func (f *WaitReadInterceptorFactory) SetWaitPerByte(w time.Duration) {
	f.mu.Lock()
	f.waitPerByte = w
	f.mu.Unlock()
}

// SetMaxChunkSize sets maximum size of the uploading data chunk.
// It can be changed during upload.
func (f *WaitReadInterceptorFactory) SetMaxChunkSize(s int) {
	f.mu.Lock()
	f.maxChunkSize = s
	f.mu.Unlock()
}

type waitReadInterceptor struct {
	factory *WaitReadInterceptorFactory
}

// New creates WaitReadInterceptor.
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
