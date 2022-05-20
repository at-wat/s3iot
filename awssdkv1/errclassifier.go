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

package awssdkv1

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
)

// ErrorClassifier classifies aws-sdk-go (v1) errors.
type ErrorClassifier struct {
	ThrottleWait time.Duration
}

// DefaultThrottleWait is a default wait duration on throttle.
var DefaultThrottleWait = 5 * time.Second

// IsRetryable implements ErrorClassifier.
func (ErrorClassifier) IsRetryable(err error) bool {
	if request.IsErrorRetryable(err) || request.IsErrorThrottle(err) {
		return true
	}
	// Workaround https://github.com/aws/aws-sdk-go/issues/3971
	if strings.Contains(err.Error(), "read: connection reset") {
		return true
	}
	return false
}

// IsThrottle implements ErrorClassifier.
func (c ErrorClassifier) IsThrottle(err error) (time.Duration, bool) {
	if !request.IsErrorThrottle(err) {
		return 0, false
	}
	wait := c.ThrottleWait
	if wait == 0 {
		wait = DefaultThrottleWait
	}
	return wait, request.IsErrorThrottle(err)
}
