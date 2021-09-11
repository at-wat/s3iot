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

package rng

import (
	"errors"
	"reflect"
	"testing"
)

func TestRange(t *testing.T) {
	t.Run("Stringer", func(t *testing.T) {
		r := Range{
			Unit:  RangeUnitBytes,
			Start: 10,
			End:   31,
			Size:  12345,
		}
		s := r.String()
		expected := "bytes=10-31"
		if s != expected {
			t.Errorf("Expected: '%s', got: '%s'", expected, s)
		}
	})
	t.Run("ParseContentRange", func(t *testing.T) {
		testCases := map[string]struct {
			input    string
			err      error
			expected Range
		}{
			"Full": {
				input: "bytes 20-30/100",
				expected: Range{
					Unit:  RangeUnitBytes,
					Start: 20,
					End:   30,
					Size:  100,
				},
			},
			"UnknownSize": {
				input: "bytes 20-30/*",
				expected: Range{
					Unit:  RangeUnitBytes,
					Start: 20,
					End:   30,
					Size:  -1,
				},
			},
			"UnknownRange": {
				input: "bytes */100",
				expected: Range{
					Unit:  RangeUnitBytes,
					Start: -1,
					End:   -1,
					Size:  100,
				},
			},
			"InvalidFormat": {
				input: "bytes=1,10/100",
				err:   ErrInvalidFormat,
			},
			"InvalidUnit": {
				input: "meters 1,10/100",
				err:   ErrInvalidUnit,
			},
			"InvalidRange": {
				input: "bytes 1,10\\100",
				err:   ErrInvalidRange,
			},
			"InvalidRange2": {
				input: "bytes 1.10/100",
				err:   ErrInvalidRange,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				r, err := ParseContentRange(tt.input)
				if err != nil {
					if tt.err == nil {
						t.Fatal(err)
					}
					if !errors.Is(err, tt.err) {
						t.Fatalf("Expected error: '%v', got: '%v'", tt.err, err)
					}
					return
				}
				if !reflect.DeepEqual(tt.expected, *r) {
					t.Errorf("Expected: %+v, got: %+v", tt.expected, *r)
				}
			})
		}
	})
}
