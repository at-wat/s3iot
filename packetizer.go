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

// DefaultPacketizerFactory is a factory of the default packetization logic.
type DefaultPacketizerFactory struct {
	PartSize       int64
	MaxUploadParts int
}

// New creates Packetizer for the given io.Reader.
func (f DefaultPacketizerFactory) New(r io.Reader) (Packetizer, error) {
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
			return &defaultPacketizerSingle{
				r:    r,
				size: size,
			}, nil
		}
		switch r := r.(type) {
		case readAtSeeker:
			return &defaultPacketizerMultiAtSeeker{
				factory: f,
				r:       r,
				size:    size,
			}, nil
		default:
		}
	default:
		size = -1
	}
	return &defaultPacketizerMulti{
		r:    r,
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, f.PartSize)
			},
		},
	}, nil
}

type defaultPacketizerSingle struct {
	r    io.ReadSeeker
	size int64
}

func (p *defaultPacketizerSingle) NextReader() (io.ReadSeeker, func(), error) {
	return p.r, func() {}, io.EOF
}

func (p *defaultPacketizerSingle) Len() int64 {
	return p.size
}

type readAtSeeker interface {
	io.ReadSeeker
	io.ReaderAt
}

type defaultPacketizerMultiAtSeeker struct {
	factory DefaultPacketizerFactory
	r       readAtSeeker
	size    int64
	offset  int64
}

func (p *defaultPacketizerMultiAtSeeker) NextReader() (io.ReadSeeker, func(), error) {
	size := p.factory.PartSize
	if p.offset+size >= p.size {
		size = p.size - p.offset
	}
	r := io.NewSectionReader(p.r, p.offset, size)
	p.offset += p.factory.PartSize
	var err error
	if p.offset >= p.size {
		err = io.EOF
	}
	return r, func() {}, err
}

func (p *defaultPacketizerMultiAtSeeker) Len() int64 {
	return p.size
}

type defaultPacketizerMulti struct {
	r      io.Reader
	size   int64
	offset int64
	pool   sync.Pool
}

func (p *defaultPacketizerMulti) NextReader() (io.ReadSeeker, func(), error) {
	buf := p.pool.Get().([]byte)
	n, err := io.ReadFull(p.r, buf)
	switch {
	case err == io.ErrUnexpectedEOF:
		err = io.EOF
	case err != nil:
		return nil, nil, err
	}
	p.offset += int64(n)
	return bytes.NewReader(buf[:n]), func() {
		p.pool.Put(buf)
	}, err
}

func (p *defaultPacketizerMulti) Len() int64 {
	return -1
}
