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

package iotest

import (
	"io"
	"sync/atomic"
)

// SeekErrorer injects seek error.
type SeekErrorer struct {
	io.ReadSeeker
	Errs []error

	cnt int32
}

// Seek returns error.
func (s *SeekErrorer) Seek(o int64, w int) (int64, error) {
	i := atomic.AddInt32(&s.cnt, 1)
	if len(s.Errs) > int(i-1) && s.Errs[i-1] != nil {
		return 0, s.Errs[i-1]
	}
	return s.Seek(o, w)
}
