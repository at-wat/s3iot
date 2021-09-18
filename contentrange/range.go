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

// Package contentrange handles HTTP Range header.
package contentrange

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Exposed errors.
var (
	ErrInvalidFormat = errors.New("invalid range header format")
	ErrInvalidUnit   = errors.New("invalid range unit")
	ErrInvalidRange  = errors.New("invalid range format")
)

// Range stores content range.
type Range struct {
	Unit  RangeUnit
	Start int64
	End   int64
	Size  int64
}

// RangeUnit represents type of range specifier.
type RangeUnit string

// RangeUnit consts.
const (
	RangeUnitBytes = "bytes"
)

// Validate range unit.
func (u RangeUnit) Validate() error {
	switch u {
	case RangeUnitBytes:
		return nil
	default:
		return fmt.Errorf("%s: %w", u, ErrInvalidUnit)
	}
}

// String returns data in HTML Range request header format.
func (r Range) String() string {
	return fmt.Sprintf("%s=%d-%d", r.Unit, r.Start, r.End)
}

// ContentRange returns data in HTML ContentRange header format.
func (r Range) ContentRange() string {
	return fmt.Sprintf("%s %d-%d/%d", r.Unit, r.Start, r.End, r.Size)
}

// Parse HTML Range request header.
func Parse(s string) (*Range, error) {
	ur := strings.Split(s, "=")
	if len(ur) != 2 {
		return nil, ErrInvalidFormat
	}
	r := &Range{
		Unit: RangeUnit(ur[0]),
	}
	if err := r.Unit.Validate(); err != nil {
		return nil, err
	}
	var err error
	se := strings.Split(ur[1], "-")
	if len(se) != 2 {
		return nil, ErrInvalidRange
	}
	if r.Start, err = strconv.ParseInt(se[0], 10, 64); err != nil {
		return nil, fmt.Errorf("content range start: %w", err)
	}
	if r.End, err = strconv.ParseInt(se[1], 10, 64); err != nil {
		return nil, fmt.Errorf("content range end: %w", err)
	}
	return r, nil
}

// ParseContentRange parses HTML Content-Range header.
func ParseContentRange(s string) (*Range, error) {
	ur := strings.Fields(s)
	if len(ur) != 2 {
		return nil, ErrInvalidFormat
	}
	r := &Range{
		Unit: RangeUnit(ur[0]),
	}
	if err := r.Unit.Validate(); err != nil {
		return nil, err
	}
	rs := strings.Split(ur[1], "/")
	if len(rs) != 2 {
		return nil, ErrInvalidRange
	}
	if rs[1] == "*" {
		r.Size = -1
	} else {
		var err error
		if r.Size, err = strconv.ParseInt(rs[1], 10, 64); err != nil {
			return nil, fmt.Errorf("content range size: %w", err)
		}
	}
	if rs[0] == "*" {
		r.Start = -1
		r.End = -1
	} else {
		var err error
		se := strings.Split(rs[0], "-")
		if len(se) != 2 {
			return nil, ErrInvalidRange
		}
		if r.Start, err = strconv.ParseInt(se[0], 10, 64); err != nil {
			return nil, fmt.Errorf("content range start: %w", err)
		}
		if r.End, err = strconv.ParseInt(se[1], 10, 64); err != nil {
			return nil, fmt.Errorf("content range end: %w", err)
		}
	}
	return r, nil
}
