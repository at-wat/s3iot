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
	"io"
)

type atWriter struct {
	w      io.WriterAt
	offset int64
}

func (w *atWriter) Write(b []byte) (int, error) {
	n, err := w.w.WriteAt(b, w.offset)
	w.offset += int64(n)
	return n, err
}

func (w *atWriter) WriteAt(b []byte, offset int64) (int, error) {
	n, err := w.w.WriteAt(b, w.offset+offset)
	return n, err
}
