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
	"bytes"
	"io"
	"testing"
)

func TestDefaultUploadSlicer(t *testing.T) {
	testCases := map[string]struct {
		partSize int64
		input    []byte
		expected [][]byte
	}{
		"Single": {
			partSize: 64,
			input:    []byte("0123456789abcdef"),
			expected: [][]byte{[]byte("0123456789abcdef")},
		},
		"Multi": {
			partSize: 5,
			input:    []byte("0123456789abcdef"),
			expected: [][]byte{
				[]byte("01234"),
				[]byte("56789"),
				[]byte("abcde"),
				[]byte("f"),
			},
		},
		"MultiAligned": {
			partSize: 4,
			input:    []byte("0123456789abcdef"),
			expected: [][]byte{
				[]byte("0123"),
				[]byte("4567"),
				[]byte("89ab"),
				[]byte("cdef"),
			},
		},
		"DefaultParam": {
			input:    []byte("0123456789abcdef"),
			expected: [][]byte{[]byte("0123456789abcdef")},
		},
	}
	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			f := &DefaultUploadSlicerFactory{
				PartSize: tt.partSize,
			}
			testCases := map[string]struct {
				r io.Reader
				n int64
			}{
				"Reader": {
					r: &readOnly{
						r: bytes.NewReader(tt.input),
					},
					n: -1,
				},
				"ReadSeeker": {
					r: &readSeekOnly{
						r: bytes.NewReader(tt.input),
					},
					n: int64(len(tt.input)),
				},
				"ReadSeekerAt": {
					r: bytes.NewReader(tt.input),
					n: int64(len(tt.input)),
				},
			}

			for name, tt2 := range testCases {
				tt2 := tt2
				t.Run(name, func(t *testing.T) {
					s, err := f.New(tt2.r)
					if err != nil {
						t.Fatal(err)
					}
					if n := s.Len(); n != tt2.n {
						t.Errorf("UploadSlicer reported wrong length. Expected %d, got %d", tt2.n, n)
					}
					for _, e := range tt.expected {
						r, cleanup, err := s.NextReader()

						b, err := io.ReadAll(r)
						if err != nil {
							t.Fatal(err)
						}
						cleanup()
						if !bytes.Equal(e, b) {
							t.Errorf("Expected: %v, got: %v", e, b)
						}
					}
				})
			}
		})
	}
}

type readOnly struct {
	r io.Reader
}

func (r *readOnly) Read(b []byte) (int, error) {
	return r.r.Read(b)
}

type readSeekOnly struct {
	r io.ReadSeeker
}

func (r *readSeekOnly) Read(b []byte) (int, error) {
	return r.r.Read(b)
}

func (r *readSeekOnly) Seek(offset int64, whence int) (int64, error) {
	return r.r.Seek(offset, whence)
}
