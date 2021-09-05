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
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// Default retry parameters.
const (
	DefaultExponentialBackoffWaitBase = time.Second
	DefaultExponentialBackoffWaitMax  = time.Minute
	DefaultRetryMax                   = 8
)

// DefaultRetryer is the default Retryer used by Uploader.
var DefaultRetryer = &ExponentialBackoffRetryerFactory{}

// NoRetryerFactory disables retry.
type NoRetryerFactory struct{}

// New creates NoRetryer.
func (NoRetryerFactory) New() Retryer {
	return &noRetryer{}
}

type noRetryer struct{}

func (noRetryer) OnFail(int64, error) bool {
	return false
}

func (noRetryer) OnSuccess(int64) {}

// ExponentialBackoffRetryerFactory creates ExponentialBackoffRetryer.
// When raw s3 upload API call is failed, the API call will be retried
// after WaitBase. Wait duration is multiplyed by 2 if it continuously
// failed up to WaitMax.
type ExponentialBackoffRetryerFactory struct {
	WaitBase time.Duration
	WaitMax  time.Duration
	RetryMax int
}

// New creates ExponentialBackoffRetryer.
func (f ExponentialBackoffRetryerFactory) New() Retryer {
	if f.WaitBase == 0 {
		f.WaitBase = DefaultExponentialBackoffWaitBase
	}
	if f.WaitMax == 0 {
		f.WaitMax = DefaultExponentialBackoffWaitMax
	}
	if f.RetryMax == 0 {
		f.RetryMax = DefaultRetryMax
	}
	return &exponentialBackoffRetryer{
		factory: f,
		wait:    make(map[int64]time.Duration),
		fails:   make(map[int64]int),
	}
}

type exponentialBackoffRetryer struct {
	factory ExponentialBackoffRetryerFactory
	mu      sync.Mutex
	wait    map[int64]time.Duration
	fails   map[int64]int
}

func (r *exponentialBackoffRetryer) OnFail(id int64, err error) bool {
	var netErr net.Error
	if ok := errors.As(err, &netErr); !ok || !netErr.Temporary() {
		return false
	}

	var wait time.Duration
	r.mu.Lock()
	if _, ok := r.wait[id]; !ok {
		r.wait[id] = r.factory.WaitBase
		wait = r.factory.WaitBase
	} else {
		wait = r.wait[id] * 2
		if wait > r.factory.WaitMax {
			wait = r.factory.WaitMax
		}
		r.wait[id] = wait
	}
	r.fails[id]++
	cnt := r.fails[id]
	r.mu.Unlock()

	if cnt > r.factory.RetryMax {
		return false
	}

	time.Sleep(wait)
	return true
}

func (r *exponentialBackoffRetryer) OnSuccess(id int64) {
	r.mu.Lock()
	if _, ok := r.wait[id]; ok {
		delete(r.wait, id)
		delete(r.fails, id)
	}
	r.mu.Unlock()
}

func withRetry(ctx context.Context, id int64, retryer Retryer, fn func() error) error {
	for {
		err := fn()
		if err != nil {
			if ctx.Err() == nil && retryer.OnFail(id, err) {
				continue
			}
			return err
		}
		retryer.OnSuccess(id)
		return nil
	}
}
