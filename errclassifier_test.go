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
	"testing"
)

var _ ErrorClassifier = &NaiveErrorClassifier{}

func TestNaiveErrorClassifier(t *testing.T) {
	c := &NaiveErrorClassifier{}
	if !c.IsRetryable(nil) {
		t.Error("IsRetryable should always return true")
	}
	if d, ok := c.IsThrottle(nil); d != 0 || ok {
		t.Errorf("IsThrottle should return (0, false), got (%v, %v)", d, ok)
	}
}
