package s3iot_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/at-wat/s3iot"
	mock_s3iot "github.com/at-wat/s3iot/internal/moq/s3iot"
)

func TestUploader(t *testing.T) {
	var (
		bucket = "Bucket"
		key    = "Key"
	)
	data := make([]byte, 128)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	t.Run("SinglePart", func(t *testing.T) {
		testCases := map[string]struct {
			num   map[string]int
			err   error
			calls int
		}{
			"NoAPIError": {
				num:   map[string]int{},
				calls: 1,
			},
			"OneAPIError": {
				num:   map[string]int{"put": 1},
				calls: 2,
			},
			"TwoPutAPIError": {
				num: map[string]int{"put": 2},
				err: errTemp,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				api := newMockAPI(buf, tt.num)
				u := &s3iot.Uploader{}
				s3iot.WithAPI(api)(u)
				s3iot.WithPacketizer(&s3iot.DefaultPacketizerFactory{PartSize: 128})(u)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				})(u)
				s3iot.WithReadInterceptor(nil)(u)

				uc, err := u.Upload(context.TODO(), &s3iot.UploadInput{
					Bucket: &bucket,
					Key:    &key,
					Body:   bytes.NewReader(data),
				})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-time.After(time.Second):
					t.Fatal("Timeout")
				case <-uc.Done():
				}
				out, err := uc.Result()
				if tt.err == nil {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					if !errors.Is(err, errTemp) {
						t.Fatalf("Expected error: '%v', got: '%v'", errTemp, err)
					}
					if n := len(api.AbortMultipartUploadCalls()); n != 1 {
						t.Fatalf("AbortMultipartUpload must be called once, but called %d times", n)
					}
					return
				}

				if len(api.CreateMultipartUploadCalls()) != 0 {
					t.Fatal("CreateMultipartUpload must not be called")
				}
				if n := len(api.PutObjectCalls()); n != tt.calls {
					t.Fatalf("PutObject must be called %d times, but called %d times", tt.calls, n)
				}

				if *out.ETag != "TAG1" {
					t.Errorf("Expected ETag: TAG0, got: %s", *out.ETag)
				}
				if !bytes.Equal(data, buf.Bytes()) {
					t.Error("Uploaded data differs")
				}
			})
		}
	})
	t.Run("MultiPart", func(t *testing.T) {
		testCases := map[string]struct {
			num   map[string]int
			err   error
			calls int
		}{
			"NoAPIError": {
				num:   map[string]int{},
				calls: 1,
			},
			"OneAPIError": {
				num:   map[string]int{"complete": 1, "create": 1, "upload": 1},
				calls: 2,
			},
			"TwoCompleteAPIError": {
				num: map[string]int{"complete": 2},
				err: errTemp,
			},
			"TwoCreateAPIError": {
				num: map[string]int{"create": 2},
				err: errTemp,
			},
			"TwoUploadAPIError": {
				num: map[string]int{"upload": 2},
				err: errTemp,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				api := newMockAPI(buf, tt.num)
				u := &s3iot.Uploader{}
				s3iot.WithAPI(api)(u)
				s3iot.WithPacketizer(&s3iot.DefaultPacketizerFactory{PartSize: 50})(u)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				})(u)
				s3iot.WithReadInterceptor(nil)(u)

				uc, err := u.Upload(context.TODO(), &s3iot.UploadInput{
					Bucket: &bucket,
					Key:    &key,
					Body:   bytes.NewReader(data),
				})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-time.After(time.Second):
					t.Fatal("Timeout")
				case <-uc.Done():
				}
				out, err := uc.Result()
				if tt.err == nil {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					if !errors.Is(err, errTemp) {
						t.Fatalf("Expected error: '%v', got: '%v'", errTemp, err)
					}
					if n := len(api.AbortMultipartUploadCalls()); n != 1 {
						t.Fatalf("AbortMultipartUpload must be called once, but called %d times", n)
					}
					return
				}

				if n := len(api.CreateMultipartUploadCalls()); n != tt.calls {
					t.Fatalf("CreateMultipartUpload must be called %d times, but called %d times", tt.calls, n)
				}
				if n := len(api.UploadPartCalls()); n != 3+tt.num["upload"] {
					t.Fatalf("UploadPart must be called %d times, but called %d times", 3+tt.num["upload"], n)
				}
				if n := len(api.CompleteMultipartUploadCalls()); n != tt.calls {
					t.Fatalf("CompleteMultipartUpload must be called %d times, but called %d times", tt.calls, n)
				}
				if len(api.PutObjectCalls()) != 0 {
					t.Fatal("PutObject must not be called")
				}

				if *out.ETag != "TAG4" {
					t.Errorf("Expected ETag: TAG0, got: %s", *out.ETag)
				}
				if !bytes.Equal(data, buf.Bytes()) {
					t.Error("Uploaded data differs")
				}

				comp := api.CompleteMultipartUploadCalls()
				if n := len(comp[0].Input.CompletedParts); n != 3 {
					t.Fatalf("Expected 3 parts, actually %d parts", n)
				}
				for i, expected := range []string{"TAG1", "TAG2", "TAG3"} {
					tag := comp[0].Input.CompletedParts[i].ETag
					if expected != *tag {
						t.Errorf("Part %d must have ETag: %s, actual: %s", i, expected, *tag)
					}
				}
			})
		}
	})
}

func newMockAPI(buf *bytes.Buffer, num map[string]int) *mock_s3iot.MockS3API {
	var etag int32
	genETag := func() *string {
		i := atomic.AddInt32(&etag, 1)
		s := fmt.Sprintf("TAG%d", i)
		return &s
	}
	uploadID := "UPLOAD0"

	var mu sync.Mutex
	cnt := make(map[string]int)
	count := func(name string) int {
		mu.Lock()
		count := cnt[name]
		cnt[name]++
		mu.Unlock()
		return count
	}

	return &mock_s3iot.MockS3API{
		AbortMultipartUploadFunc: func(ctx context.Context, input *s3iot.AbortMultipartUploadInput) (*s3iot.AbortMultipartUploadOutput, error) {
			if count("abort") < num["abort"] {
				return nil, errTemp
			}
			return &s3iot.AbortMultipartUploadOutput{}, nil
		},
		CompleteMultipartUploadFunc: func(ctx context.Context, input *s3iot.CompleteMultipartUploadInput) (*s3iot.CompleteMultipartUploadOutput, error) {
			if count("complete") < num["complete"] {
				return nil, errTemp
			}
			return &s3iot.CompleteMultipartUploadOutput{
				ETag: genETag(),
			}, nil
		},
		CreateMultipartUploadFunc: func(ctx context.Context, input *s3iot.CreateMultipartUploadInput) (*s3iot.CreateMultipartUploadOutput, error) {
			if count("create") < num["create"] {
				return nil, errTemp
			}
			return &s3iot.CreateMultipartUploadOutput{
				UploadID: &uploadID,
			}, nil
		},
		PutObjectFunc: func(ctx context.Context, input *s3iot.PutObjectInput) (*s3iot.PutObjectOutput, error) {
			if count("put") < num["put"] {
				return nil, errTemp
			}
			io.Copy(buf, input.Body)
			return &s3iot.PutObjectOutput{
				ETag: genETag(),
			}, nil
		},
		UploadPartFunc: func(ctx context.Context, input *s3iot.UploadPartInput) (*s3iot.UploadPartOutput, error) {
			if count("upload") < num["upload"] {
				return nil, errTemp
			}
			io.Copy(buf, input.Body)
			return &s3iot.UploadPartOutput{
				ETag: genETag(),
			}, nil
		},
	}
}

type errTemporary struct{}

func (errTemporary) Temporary() bool { return true }
func (errTemporary) Timeout() bool   { return true }
func (errTemporary) Error() string   { return "timeout" }

var errTemp = &errTemporary{}
