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

func TestWithRetry(t *testing.T) {
	f := &ExponentialBackoffRetryerFactory{
		WaitBase: 1 * time.Millisecond,
		RetryMax: 1,
	}

	t.Run("Success", func(t *testing.T) {
		r := f.New(nil)
		var i int
		err := withRetry(context.TODO(), 0, r, &NaiveErrorClassifier{}, func() error {
			defer func() {
				i++
			}()
			switch i {
			case 0:
				return nil
			default:
				t.Fatal("Must not reach here")
				return nil
			}
		})
		if err != nil {
			t.Error("Unexpected failure")
		}
		if i != 1 {
			t.Errorf("Expected retry count: 1, actual: %d", i)
		}
	})
	t.Run("SuccessAfterRetry", func(t *testing.T) {
		r := f.New(nil)
		var i int
		err := withRetry(context.TODO(), 0, r, &NaiveErrorClassifier{}, func() error {
			defer func() {
				i++
			}()
			switch i {
			case 0:
				return errDummy
			case 1:
				return nil
			default:
				t.Fatal("Must not reach here")
				return nil
			}
		})
		if err != nil {
			t.Error("Unexpected failure")
		}
		if i != 2 {
			t.Errorf("Expected retry count: 2, actual: %d", i)
		}
	})
	t.Run("Failure", func(t *testing.T) {
		r := f.New(nil)
		var i int
		err := withRetry(context.TODO(), 0, r, &NaiveErrorClassifier{}, func() error {
			defer func() {
				i++
			}()
			switch i {
			case 0, 1:
				return errDummy
			default:
				t.Fatal("Must not reach here")
				return nil
			}
		})
		if err != errDummy {
			t.Errorf("Expected error: %v, got: %v", errDummy, err)
		}
		if i != 2 {
			t.Errorf("Expected retry count: 2, actual: %d", i)
		}
	})
	t.Run("WithErrorClassifier", func(t *testing.T) {
		errRetryable := errors.New("retryable")
		errNotRetryable := errors.New("non retryable")

		ec := &dummyErrorClassifier{
			retryable: errRetryable,
		}
		t.Run("Retryable", func(t *testing.T) {
			r := f.New(nil)
			var i int
			err := withRetry(context.TODO(), 0, r, ec, func() error {
				defer func() {
					i++
				}()
				switch i {
				case 0, 1:
					return errRetryable
				default:
					t.Fatal("Must not reach here")
					return nil
				}
			})
			if err != errRetryable {
				t.Errorf("Expected error: %v, got: %v", errRetryable, err)
			}
			if i != 2 {
				t.Errorf("Expected retry count: 2, actual: %d", i)
			}
		})
		t.Run("NotRetryable", func(t *testing.T) {
			r := f.New(nil)
			var i int
			err := withRetry(context.TODO(), 0, r, ec, func() error {
				defer func() {
					i++
				}()
				switch i {
				case 0:
					return errNotRetryable
				default:
					t.Fatal("Must not reach here")
					return nil
				}
			})
			if err != errNotRetryable {
				t.Errorf("Expected error: %v, got: %v", errNotRetryable, err)
			}
			if i != 1 {
				t.Errorf("Expected retry count: 2, actual: %d", i)
			}
		})
		t.Run("CancelDuringThrottle", func(t *testing.T) {
			ec := &dummyErrorClassifier{
				retryable:    errRetryable,
				throttleWait: time.Second,
			}
			r := f.New(nil)
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			go func() {
				select {
				case <-ctx.Done():
				case <-time.After(100 * time.Millisecond):
					panic("timeout")
				}
			}()

			err := withRetry(ctx, 0, r, ec, func() error {
				return errRetryable
			})
			if err != context.DeadlineExceeded {
				t.Errorf("Expected error: %v, got: %v", context.DeadlineExceeded, err)
			}
		})
	})
}

type dummyErrorClassifier struct {
	retryable    error
	throttleWait time.Duration
}

func (ec dummyErrorClassifier) IsRetryable(err error) bool {
	return err == ec.retryable
}

func (ec dummyErrorClassifier) IsThrottle(err error) (time.Duration, bool) {
	if err == ec.retryable {
		if ec.throttleWait == 0 {
			return time.Millisecond, true
		}
		return ec.throttleWait, true
	}
	return 0, false
}
