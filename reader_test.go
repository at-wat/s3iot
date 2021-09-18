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
	"fmt"
	"io"
	"testing"
	"time"
)

func TestWaitReadInterceptor(t *testing.T) {
	f := NewWaitReadInterceptorFactory(
		2*time.Millisecond,
		WaitReadInterceptorMaxChunkSize(8),
	)

	if f.maxChunkSize != 8 {
		t.Errorf("MaxChunkSize is not configured. Expected 8, got %d", f.maxChunkSize)
	}

	f.SetMaxChunkSize(16)
	if f.maxChunkSize != 16 {
		t.Errorf("MaxChunkSize is not configured. Expected 16, got %d", f.maxChunkSize)
	}

	ri := f.New()

	const tolerance = 50 * time.Millisecond

	// Use slice to run in particular order
	testCases := []struct {
		name                string
		setWaitPerByte      time.Duration
		expectedWaitPerByte time.Duration
	}{
		{
			name:                "Argument",
			expectedWaitPerByte: 2 * time.Millisecond,
		},
		{
			name:                "SetWaitPerByte",
			setWaitPerByte:      4 * time.Millisecond,
			expectedWaitPerByte: 4 * time.Millisecond,
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.setWaitPerByte != 0 {
				f.SetWaitPerByte(tt.setWaitPerByte)
			}
			for _, n := range []int{128, 256} {
				n := n
				t.Run(fmt.Sprintf("%dBytes", n), func(t *testing.T) {
					r := bytes.NewReader(make([]byte, n))
					r2 := ri.Reader(r)
					ts := time.Now()
					if _, err := io.ReadAll(r2); err != nil {
						t.Fatal(err)
					}
					te := time.Now()

					expected := time.Duration(n) * tt.expectedWaitPerByte
					diff := te.Sub(ts) - expected
					if diff < -tolerance || tolerance < diff {
						t.Errorf("Expected duration: %v, actual: %v", expected, te.Sub(ts))
					}
				})
			}
		})
	}
}
