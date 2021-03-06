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
)

// ReadOnly removes Seek and ReadAt from Reader.
type ReadOnly struct {
	R io.Reader
}

// Read implements io.Reader.
func (r *ReadOnly) Read(b []byte) (int, error) {
	return r.R.Read(b)
}

// ReadSeekOnly removes ReadAt from Reader.
type ReadSeekOnly struct {
	R io.ReadSeeker
}

// Read implements io.Reader.
func (r *ReadSeekOnly) Read(b []byte) (int, error) {
	return r.R.Read(b)
}

// Seek implements io.Seeker.
func (r *ReadSeekOnly) Seek(offset int64, whence int) (int64, error) {
	return r.R.Seek(offset, whence)
}

// ReadErrorer injects read error.
type ReadErrorer struct {
	Err error
}

// Read implements io.Reader.
func (r ReadErrorer) Read([]byte) (int, error) {
	return 0, r.Err
}
