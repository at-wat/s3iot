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
	"testing"
)

func TestRetryError(t *testing.T) {
	base := errors.New("base")
	err := RetryError{error: base}

	expected := "retry exceeded limit: base"
	if s := err.Error(); s != expected {
		t.Errorf("Expected: '%s', got: '%s'", expected, s)
	}
	if ue := err.Unwrap(); ue != base {
		t.Errorf("Expected unwrapped error: '%v', got: '%v'", base, ue)
	}
}
