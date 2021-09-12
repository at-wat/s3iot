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
	"reflect"
	"testing"

	"github.com/at-wat/s3iot/internal/iotest"
	"github.com/at-wat/s3iot/rng"
)

func TestDefaultDownloadSlicer(t *testing.T) {
	testCases := map[string]struct {
		partSize int64
		bufSize  int
		write    [][]byte
		ranges   []rng.Range
		expected []byte
	}{
		"Dense": {
			partSize: 4,
			bufSize:  16,
			write: [][]byte{
				[]byte("0123"),
				[]byte("4567"),
				[]byte("89ab"),
				[]byte("cdef"),
			},
			ranges: []rng.Range{
				{
					Unit:  rng.RangeUnitBytes,
					Start: 0,
					End:   3,
				},
				{
					Unit:  rng.RangeUnitBytes,
					Start: 4,
					End:   7,
				},
				{
					Unit:  rng.RangeUnitBytes,
					Start: 8,
					End:   11,
				},
				{
					Unit:  rng.RangeUnitBytes,
					Start: 12,
					End:   15,
				},
			},
			expected: []byte("0123456789abcdef"),
		},
		"Sparse": {
			partSize: 5,
			bufSize:  16,
			write: [][]byte{
				[]byte("0123"),
				[]byte("4567"),
				[]byte("89ab"),
			},
			ranges: []rng.Range{
				{
					Unit:  rng.RangeUnitBytes,
					Start: 0,
					End:   4,
				},
				{
					Unit:  rng.RangeUnitBytes,
					Start: 5,
					End:   9,
				},
				{
					Unit:  rng.RangeUnitBytes,
					Start: 10,
					End:   14,
				},
			},
			expected: []byte("0123\0004567\00089ab\000\000"),
		},
		"DefaultParam": {
			bufSize: 5,
			write:   [][]byte{[]byte("0123")},
			ranges: []rng.Range{
				{
					Unit:  rng.RangeUnitBytes,
					Start: 0,
					End:   5242879,
				},
			},
			expected: []byte("0123\000"),
		},
	}
	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			f := &DefaultDownloadSlicerFactory{
				PartSize: tt.partSize,
			}
			buf := iotest.BufferAt(make([]byte, tt.bufSize))
			s, err := f.New(buf)
			if err != nil {
				t.Fatal(err)
			}
			for i, data := range tt.write {
				w, r, err := s.NextWriter()
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(tt.ranges[i], r) {
					t.Errorf("Expected range: %v, got: %v", tt.ranges[i], r)
				}

				_, _ = w.WriteAt(data, 0)
			}
			if !bytes.Equal(tt.expected, []byte(buf)) {
				t.Errorf("Expected: %v, got: %v", tt.expected, []byte(buf))
			}
		})
	}
}
