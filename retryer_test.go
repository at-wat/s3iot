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
	"testing"
	"time"
)

var _ RetryerFactory = &NoRetryerFactory{}
var _ RetryerFactory = &ExponentialBackoffRetryerFactory{}
var _ RetryerFactory = &PauseOnFailRetryerFactory{}
var _ RetryerFactory = &RetryerHookFactory{}

var errDummy = errors.New("dummy")

func TestNoRetryer(t *testing.T) {
	f := &NoRetryerFactory{}
	r := f.New(nil)
	if cont := r.OnFail(context.TODO(), 0, nil); cont {
		t.Error("Unexpected retry")
	}
	r.OnSuccess(0)
}

func TestExponentialBackoffRetryer(t *testing.T) {
	f := &ExponentialBackoffRetryerFactory{
		WaitBase: 50 * time.Millisecond,
		WaitMax:  250 * time.Millisecond,
		RetryMax: 4,
	}
	r := f.New(nil)

	ts := time.Now()
	for i := 0; i < 4; i++ {
		if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
			t.Error("Unexpected failure before reaching RetryMax")
		}
	}
	te := time.Now()
	// 50 + 100 + 200 + 250 = 600ms
	expectedWait := 600 * time.Millisecond
	tolerance := 50 * time.Millisecond
	if diff := te.Sub(ts) - expectedWait; diff < -tolerance || tolerance < diff {
		t.Errorf("Expected wait: %v, actual: %v", te.Sub(ts), expectedWait)
	}

	if cont := r.OnFail(context.TODO(), 1, errDummy); !cont {
		t.Error("Unexpected failure on different id")
	}

	if cont := r.OnFail(context.TODO(), 0, errDummy); cont {
		t.Error("Unexpected retry after reaching RetryMax")
	}
	r.OnSuccess(0)

	if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
		t.Error("Unexpected failure after resetting failure")
	}
}

func TestPauseOnFailRetryer(t *testing.T) {
	t.Run("WithExponentialBackoffRetryer", func(t *testing.T) {
		f := &PauseOnFailRetryerFactory{
			Base: &ExponentialBackoffRetryerFactory{
				WaitBase: 50 * time.Millisecond,
				WaitMax:  50 * time.Millisecond,
				RetryMax: 1,
			},
		}
		uc := &dummyUploadContext{
			chPause: make(chan struct{}, 1),
		}
		r := f.New(uc)

		if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
			t.Error("Unexpected failure before reaching RetryMax")
		}
		if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
			t.Error("PauseOnFailRetryer should not abort")
		}

		select {
		case <-uc.chPause:
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		}

		r.OnSuccess(0)

		if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
			t.Error("PauseOnFailRetryer should not abort")
		}
	})
	t.Run("WithNoRetryer", func(t *testing.T) {
		f := &PauseOnFailRetryerFactory{}
		uc := &dummyUploadContext{
			chPause: make(chan struct{}, 1),
		}
		r := f.New(uc)

		if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
			t.Error("PauseOnFailRetryer should not abort")
		}

		select {
		case <-uc.chPause:
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		}

		r.OnSuccess(0)

		if cont := r.OnFail(context.TODO(), 0, errDummy); !cont {
			t.Error("PauseOnFailRetryer should not abort")
		}
	})
}

func TestRetryerHookFactory(t *testing.T) {
	var bucketStored, keyStored string
	var errStored error
	f := &RetryerHookFactory{
		Base: &dummyRetryerFactory{},
		OnError: func(bucket, key string, err error) {
			if errStored != nil {
				t.Error("OnError is expected to be called once")
			}
			bucketStored, keyStored, errStored = bucket, key, err
		},
	}
	r := f.New(&dummyUploadContext{})

	if cont := r.OnFail(context.TODO(), 0, errDummy); cont {
		t.Error("Expected no-continue")
	}
	r.OnSuccess(0)

	if n := f.Base.(*dummyRetryerFactory).numFail; n != 1 {
		t.Errorf("Base retryer OnFail must be called once, but called %d times", n)
	}
	if n := f.Base.(*dummyRetryerFactory).numSuccess; n != 1 {
		t.Errorf("Base retryer OnSuccess must be called once, but called %d times", n)
	}

	if bucketStored != "dummyBucket" {
		t.Errorf("Expected dummyBucket, got %s", bucketStored)
	}
	if keyStored != "dummyKey" {
		t.Errorf("Expected dummyKey, got %s", keyStored)
	}
	if errStored != errDummy {
		t.Errorf("Expected '%v', got '%v'", errDummy, errStored)
	}
}

type dummyRetryerFactory struct {
	numFail    int
	numSuccess int
}

func (f *dummyRetryerFactory) New(Pauser) Retryer {
	return &dummyRetryer{dummyRetryerFactory: f}
}

type dummyRetryer struct {
	*dummyRetryerFactory
}

func (r *dummyRetryer) OnFail(context.Context, int64, error) bool {
	r.numFail++
	return false
}

func (r *dummyRetryer) OnSuccess(int64) {
	r.numSuccess++
}

type dummyUploadContext struct {
	UploadContext
	chPause chan struct{}
}

func (dummyUploadContext) BucketKey() (bucket, key string) {
	return "dummyBucket", "dummyKey"
}

func (uc *dummyUploadContext) Pause() {
	uc.chPause <- struct{}{}
}
