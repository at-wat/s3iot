package s3iot

import (
	"bytes"
	"io"
	"testing"
)

func TestDefaultPacketizer(t *testing.T) {
	f := &DefaultPacketizerFactory{
		PartSize: 5,
	}

	input := []byte("0123456789abcdef")
	expected := [][]byte{
		[]byte("01234"),
		[]byte("56789"),
		[]byte("abcde"),
		[]byte("f"),
	}

	readers := map[string]io.Reader{
		"Reader": &readOnly{
			r: bytes.NewReader(input),
		},
		"ReadSeeker": &readSeekOnly{
			r: bytes.NewReader(input),
		},
		"ReadSeekerAt": bytes.NewReader(input),
	}

	for name, r := range readers {
		r := r
		t.Run(name, func(t *testing.T) {
			p, err := f.New(r)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range expected {
				r, cleanup, err := p.NextReader()

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
