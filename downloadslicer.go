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

	"github.com/at-wat/s3iot/contentrange"
)

// Default download slicing parametes.
const (
	DefaultDownloadPartSize = 1024 * 1024 * 5
)

// DefaultDownloadSlicerFactory is a factory of the default slicing logic.
type DefaultDownloadSlicerFactory struct {
	PartSize int64
}

// New creates DownloadSlicer for the given io.WriterAt.
func (f DefaultDownloadSlicerFactory) New(w io.WriterAt) DownloadSlicer {
	if f.PartSize == 0 {
		f.PartSize = DefaultDownloadPartSize
	}
	return &defaultDownloadSlicer{
		factory: f,
		w:       w,
	}
}

type defaultDownloadSlicer struct {
	factory DefaultDownloadSlicerFactory
	w       io.WriterAt
	offset  int64
}

func (s *defaultDownloadSlicer) NextWriter() (io.WriterAt, contentrange.Range) {
	r := contentrange.Range{
		Unit:  contentrange.RangeUnitBytes,
		Start: s.offset,
		End:   s.offset + s.factory.PartSize - 1,
	}
	s.offset += s.factory.PartSize
	return &atWriter{w: s.w, offset: r.Start}, r
}
