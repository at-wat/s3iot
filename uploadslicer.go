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
	"sync"
)

// Default packetization parametes.
const (
	DefaultUploadPartSize = 1024 * 1024 * 5
	MaxUploadParts        = 10000
)

// DefaultUploadSlicerFactory is a factory of the default packetization logic.
type DefaultUploadSlicerFactory struct {
	PartSize       int64
	MaxUploadParts int
}

// New creates UploadSlicer for the given io.Reader.
func (f DefaultUploadSlicerFactory) New(r io.Reader) (UploadSlicer, error) {
	if f.PartSize == 0 {
		f.PartSize = DefaultUploadPartSize
	}
	if f.MaxUploadParts == 0 {
		f.MaxUploadParts = MaxUploadParts
	}
	var size int64
	switch r := r.(type) {
	case io.ReadSeeker:
		var err error
		if size, err = r.Seek(0, io.SeekEnd); err != nil {
			return nil, err
		}
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		n := size / f.PartSize
		if n == 0 {
			return &defaultUploadSlicerSingle{
				r:    r,
				size: size,
			}, nil
		}
		switch r := r.(type) {
		case readAtSeeker:
			return &defaultUploadSlicerMultiAtSeeker{
				factory: f,
				r:       r,
				size:    size,
			}, nil
		default:
		}
	default:
		size = -1
	}
	return &defaultUploadSlicerMulti{
		r:    r,
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, f.PartSize)
			},
		},
	}, nil
}

type defaultUploadSlicerSingle struct {
	r    io.ReadSeeker
	size int64
}

func (s *defaultUploadSlicerSingle) NextReader() (io.ReadSeeker, func(), error) {
	return s.r, func() {}, io.EOF
}

func (s *defaultUploadSlicerSingle) Len() int64 {
	return s.size
}

type readAtSeeker interface {
	io.ReadSeeker
	io.ReaderAt
}

type defaultUploadSlicerMultiAtSeeker struct {
	factory DefaultUploadSlicerFactory
	r       readAtSeeker
	size    int64
	offset  int64
}

func (s *defaultUploadSlicerMultiAtSeeker) NextReader() (io.ReadSeeker, func(), error) {
	size := s.factory.PartSize
	if s.offset+size >= s.size {
		size = s.size - s.offset
	}
	r := io.NewSectionReader(s.r, s.offset, size)
	s.offset += s.factory.PartSize
	var err error
	if s.offset >= s.size {
		err = io.EOF
	}
	return r, func() {}, err
}

func (s *defaultUploadSlicerMultiAtSeeker) Len() int64 {
	return s.size
}

type defaultUploadSlicerMulti struct {
	r      io.Reader
	size   int64
	offset int64
	pool   sync.Pool
}

func (s *defaultUploadSlicerMulti) NextReader() (io.ReadSeeker, func(), error) {
	buf := s.pool.Get().([]byte)
	n, err := io.ReadFull(s.r, buf)
	switch {
	case err == io.ErrUnexpectedEOF:
		err = io.EOF
	case err != nil:
		return nil, nil, err
	}
	s.offset += int64(n)
	return bytes.NewReader(buf[:n]), func() {
		s.pool.Put(buf)
	}, err
}

func (s *defaultUploadSlicerMulti) Len() int64 {
	return s.size
}
