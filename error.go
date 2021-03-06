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
	"errors"
)

// ErrForcePaused indicates part up/download is canceled due to force pause.
var ErrForcePaused = &retryableError{errors.New("force paused")}

// RetryError is returned when retry is exceeded limit.
type RetryError struct {
	error
}

// Error implements error.
func (e *RetryError) Error() string {
	return "retry exceeded limit: " + e.error.Error()
}

// Unwrap returns original error.
func (e *RetryError) Unwrap() error {
	return e.error
}
