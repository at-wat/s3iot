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

// Package bufferat provides simple io.WriterAt based on the byte array.
package bufferat

// BufferAt implements simple io.WriterAt interface.
type BufferAt []byte

// WriteAt implements io.WriterAt.
func (b BufferAt) WriteAt(p []byte, offset int64) (int, error) {
	return copy(b[int(offset):int(offset)+len(p)], p), nil
}
