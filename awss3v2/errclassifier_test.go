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

package awss3v2

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
)

func TestErrorClassifier(t *testing.T) {
	testCases := map[string]struct {
		err       error
		retryable bool
		throttle  bool
		wait      time.Duration
	}{
		"HTTP": {
			err: &http.RequestSendError{
				Err: errors.New("dummy"),
			},
			retryable: true,
		},
		"AWSRetryable": {
			err: &smithy.GenericAPIError{
				Code:    "RequestTimeout",
				Message: "dummy",
			},
			retryable: true,
		},
		"AWSNonRetryable": {
			err: &smithy.GenericAPIError{
				Code:    "DummyFatal",
				Message: "dummy",
			},
		},
		"AWSThrottle": {
			err: &smithy.GenericAPIError{
				Code:    "SlowDown",
				Message: "dummy",
			},
			retryable: true,
			throttle:  true,
			wait:      DefaultThrottleWait,
		},
	}

	ec := &ErrorClassifier{}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			if out := ec.IsRetryable(tt.err); out != tt.retryable {
				t.Errorf("IsRetryable('%v') is expected to be %v, got %v", tt.err, tt.retryable, out)
			}
			wait, throttle := ec.IsThrottle(tt.err)
			if throttle != tt.throttle {
				t.Errorf("IsThrottle('%v') is expected to be %v, got %v", tt.err, tt.throttle, throttle)
			}
			if wait != tt.wait {
				t.Errorf("Expected wait for '%v': %v, got: %v", tt.err, tt.wait, wait)
			}
		})
	}
}
