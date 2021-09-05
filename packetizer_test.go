package s3iot

import (
	"bytes"
	"io"
	"testing"
)

func TestDefaultPacketizer(t *testing.T) {
	testCases := map[string]struct {
		partSize int64
		input    []byte
		expected [][]byte
	}{
		"Single": {
			partSize: 64,
			input:    []byte("0123456789abcdef"),
			expected: [][]byte{
				[]byte("0123456789abcdef"),
			},
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
	}
	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			f := &DefaultPacketizerFactory{
				PartSize: tt.partSize,
			}
			readers := map[string]io.Reader{
				"Reader": &readOnly{
					r: bytes.NewReader(tt.input),
				},
				"ReadSeeker": &readSeekOnly{
					r: bytes.NewReader(tt.input),
				},
				"ReadSeekerAt": bytes.NewReader(tt.input),
			}

			for name, r := range readers {
				r := r
				t.Run(name, func(t *testing.T) {
					p, err := f.New(r)
					if err != nil {
						t.Fatal(err)
					}
					for _, e := range tt.expected {
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
