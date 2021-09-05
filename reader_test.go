package s3iot

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestWaitReadInterceptor(t *testing.T) {
	const waitPerByte = 2 * time.Millisecond
	f := NewWaitReadInterceptorFactory(waitPerByte)
	f.SetMaxChunkSize(16)
	ri := f.New()

	const tolerance = 50 * time.Millisecond

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

			expected := time.Duration(n) * waitPerByte
			diff := te.Sub(ts) - expected
			if diff < -tolerance || tolerance < diff {
				t.Errorf("Expected duration: %v, actual: %v", expected, te.Sub(ts))
			}
		})
	}
}
