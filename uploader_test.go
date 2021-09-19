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
	"github.com/at-wat/s3iot/internal/iotest"
	mock_s3iot "github.com/at-wat/s3iot/internal/moq/s3iot"
)

var errTemp = errors.New("dummy")

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
				api := newUploadMockAPI(buf, tt.num, nil)
				u := &s3iot.Uploader{}
				s3iot.WithAPI(api).ApplyToUploader(u)
				s3iot.WithUploadSlicer(&s3iot.DefaultUploadSlicerFactory{PartSize: 128}).ApplyToUploader(u)
				s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				}).ApplyToUploader(u)

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
				if *out.Location != "s3://url" {
					t.Errorf("Expected Location: s3://url, got: %s", *out.Location)
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
				api := newUploadMockAPI(buf, tt.num, nil)
				u := &s3iot.Uploader{}
				s3iot.WithAPI(api).ApplyToUploader(u)
				s3iot.WithUploadSlicer(&s3iot.DefaultUploadSlicerFactory{PartSize: 50}).ApplyToUploader(u)
				s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				}).ApplyToUploader(u)

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
				if *out.Location != "s3://url" {
					t.Errorf("Expected Location: s3://url, got: %s", *out.Location)
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
	t.Run("PauseResume", func(t *testing.T) {
		buf := &bytes.Buffer{}
		chUpload := make(chan interface{})
		api := newUploadMockAPI(buf, nil, map[string]chan interface{}{
			"upload": chUpload,
		})
		u := &s3iot.Uploader{}
		s3iot.WithAPI(api).ApplyToUploader(u)
		s3iot.WithUploadSlicer(
			&s3iot.DefaultUploadSlicerFactory{PartSize: 50},
		).ApplyToUploader(u)
		s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)
		s3iot.WithRetryer(nil).ApplyToUploader(u)

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
		case <-chUpload:
		}

		uc.Pause()
		go func() {
			<-chUpload
			<-chUpload
		}()

		time.Sleep(50 * time.Millisecond)
		status, err := uc.Status()
		if err != nil {
			t.Fatal(err)
		}
		if status.UploadID != "UPLOAD0" {
			t.Errorf("Expected upload ID: UPLOAD0, got: %s", status.UploadID)
		}

		select {
		case <-time.After(500 * time.Millisecond):
		case <-uc.Done():
			t.Fatal("Upload should be paused")
		}

		uc.Resume()

		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-uc.Done():
		}
		if _, err = uc.Result(); err != nil {
			t.Fatal(err)
		}

		if n := len(api.CreateMultipartUploadCalls()); n != 1 {
			t.Fatalf("CreateMultipartUpload must be called once, but called %d times", n)
		}
		if n := len(api.UploadPartCalls()); n != 3 {
			t.Fatalf("UploadPart must be called 3 times, but called %d times", n)
		}
		if n := len(api.CompleteMultipartUploadCalls()); n != 1 {
			t.Fatalf("CompleteMultipartUpload must be called once, but called %d times", n)
		}

		if !bytes.Equal(data, buf.Bytes()) {
			t.Error("Uploaded data differs")
		}
	})
	t.Run("CancelDuringPause", func(t *testing.T) {
		buf := &bytes.Buffer{}
		chUpload := make(chan interface{})
		api := newUploadMockAPI(buf, nil, map[string]chan interface{}{
			"upload": chUpload,
		})
		u := &s3iot.Uploader{}
		s3iot.WithAPI(api).ApplyToUploader(u)
		s3iot.WithUploadSlicer(
			&s3iot.DefaultUploadSlicerFactory{PartSize: 50},
		).ApplyToUploader(u)
		s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)
		s3iot.WithRetryer(nil).ApplyToUploader(u)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		uc, err := u.Upload(ctx, &s3iot.UploadInput{
			Bucket: &bucket,
			Key:    &key,
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(50 * time.Millisecond)
		uc.Pause()

		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-chUpload:
		}

		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-uc.Done():
		}
		if _, err = uc.Result(); err != context.Canceled {
			t.Fatalf("Expected error: '%v', got: '%v'", context.Canceled, err)
		}
	})
	t.Run("Unseekable", func(t *testing.T) {
		errSeekFailure := errors.New("seek error")

		testCases := map[string]struct {
			partSize             int64
			innerSeekerFailAtMax int
		}{
			"Single": {
				partSize:             150,
				innerSeekerFailAtMax: 1,
			},
			"SingleBoundary": {
				partSize:             128,
				innerSeekerFailAtMax: 1,
			},
			"Multi": {
				partSize:             50,
				innerSeekerFailAtMax: 3,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				t.Run("OuterSeeker", func(t *testing.T) {
					errsIn := []error{errSeekFailure}
					for n := 1; n <= 2; n++ {
						errs := errsIn

						u := &s3iot.Uploader{}
						s3iot.WithRetryer(&s3iot.NoRetryerFactory{}).ApplyToUploader(u)
						s3iot.WithAPI(newUploadMockAPI(&bytes.Buffer{}, nil, nil)).ApplyToUploader(u)
						s3iot.WithUploadSlicer(
							&s3iot.DefaultUploadSlicerFactory{PartSize: tt.partSize},
						).ApplyToUploader(u)
						s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)

						t.Run(fmt.Sprintf("SeekErrorAt%d", n), func(t *testing.T) {
							_, err := u.Upload(context.TODO(), &s3iot.UploadInput{
								Bucket: &bucket,
								Key:    &key,
								Body: &iotest.SeekErrorer{
									ReadSeeker: bytes.NewReader(data),
									Errs:       errs,
								},
							})
							if !errors.Is(err, errSeekFailure) {
								t.Fatalf("Expected error: '%v', got: '%v'", errSeekFailure, err)
							}
						})
						errsIn = append([]error{nil}, errsIn...)
					}
				})
				t.Run("InnerSeeker", func(t *testing.T) {
					errsIn := [][]error{{errSeekFailure}}
					for n := 1; n <= tt.innerSeekerFailAtMax; n++ {
						errs := errsIn

						u := &s3iot.Uploader{}
						s3iot.WithRetryer(&s3iot.NoRetryerFactory{}).ApplyToUploader(u)
						s3iot.WithAPI(newUploadMockAPI(&bytes.Buffer{}, nil, nil)).ApplyToUploader(u)
						s3iot.WithUploadSlicer(
							&seekErrorUploadSlicerFactory{
								DefaultUploadSlicerFactory: s3iot.DefaultUploadSlicerFactory{
									PartSize: tt.partSize,
								},
								errs: errs,
							},
						).ApplyToUploader(u)
						s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)

						t.Run(fmt.Sprintf("SeekErrorAt%d", n), func(t *testing.T) {
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

							if _, err = uc.Result(); !errors.Is(err, errSeekFailure) {
								t.Fatalf("Expected error: '%v', got: '%v'", errSeekFailure, err)
							}
						})
						errsIn = append([][]error{{nil}}, errsIn...)
					}
				})
			})
		}
	})
	t.Run("DefaultSlicer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		u := &s3iot.Uploader{}
		s3iot.WithAPI(
			newUploadMockAPI(buf, nil, nil),
		).ApplyToUploader(u)

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

		if !bytes.Equal(data, buf.Bytes()) {
			t.Error("Uploaded data differs")
		}
	})
	t.Run("ContextCanceled", func(t *testing.T) {
		for name, opt := range map[string]s3iot.UploaderOption{
			"SinglePart": s3iot.WithUploadSlicer(
				&s3iot.DefaultUploadSlicerFactory{},
			),
			"MultiPart": s3iot.WithUploadSlicer(
				&s3iot.DefaultUploadSlicerFactory{PartSize: 50},
			),
		} {
			t.Run(name, func(t *testing.T) {
				u := &s3iot.Uploader{}
				s3iot.WithAPI(
					newUploadMockAPI(&bytes.Buffer{}, nil, nil),
				).ApplyToUploader(u)
				opt.ApplyToUploader(u)
				s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				uc, err := u.Upload(ctx, &s3iot.UploadInput{
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
				if _, err = uc.Result(); err != context.Canceled {
					t.Fatalf("Expected error: '%v', got: '%v'", context.Canceled, err)
				}
			})
		}
	})
	t.Run("WithReadInterceptor", func(t *testing.T) {
		testCases := map[string]struct {
			partSize    int64
			readerCalls int
		}{
			"Single": {
				partSize:    128,
				readerCalls: 1,
			},
			"Multi": {
				partSize:    50,
				readerCalls: 3,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				api := newUploadMockAPI(buf, nil, nil)
				u := &s3iot.Uploader{}
				s3iot.WithAPI(api).ApplyToUploader(u)
				s3iot.WithUploadSlicer(
					&s3iot.DefaultUploadSlicerFactory{
						PartSize: tt.partSize,
					}).ApplyToUploader(u)

				ri := &mock_s3iot.MockReadInterceptor{
					ReaderFunc: func(readSeeker io.ReadSeeker) io.ReadSeeker {
						return readSeeker
					},
				}
				rif := &mock_s3iot.MockReadInterceptorFactory{
					NewFunc: func() s3iot.ReadInterceptor {
						return ri
					},
				}
				s3iot.WithReadInterceptor(rif).ApplyToUploader(u)

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
				if _, err := uc.Result(); err != nil {
					t.Fatal(err)
				}

				if n := len(rif.NewCalls()); n != 1 {
					t.Fatalf("New must be called once, but called %d times", n)
				}
				if n := len(ri.ReaderCalls()); n != tt.readerCalls {
					t.Fatalf("Reader must be called %d times, but called %d times", tt.readerCalls, n)
				}

				if !bytes.Equal(data, buf.Bytes()) {
					t.Error("Uploaded data differs")
				}
			})
		}
	})
}

func newUploadMockAPI(buf *bytes.Buffer, num map[string]int, ch map[string]chan interface{}) *mock_s3iot.MockS3API {
	if num == nil {
		num = make(map[string]int)
	}
	if ch == nil {
		ch = make(map[string]chan interface{})
	}

	var etag int32
	genETag := func() *string {
		i := atomic.AddInt32(&etag, 1)
		s := fmt.Sprintf("TAG%d", i)
		return &s
	}
	uploadID := "UPLOAD0"
	location := "s3://url"

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
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if count("abort") < num["abort"] {
				return nil, errTemp
			}
			if c := ch["abort"]; c != nil {
				c <- input
			}
			return &s3iot.AbortMultipartUploadOutput{}, nil
		},
		CompleteMultipartUploadFunc: func(ctx context.Context, input *s3iot.CompleteMultipartUploadInput) (*s3iot.CompleteMultipartUploadOutput, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if count("complete") < num["complete"] {
				return nil, errTemp
			}
			if c := ch["complete"]; c != nil {
				c <- input
			}
			return &s3iot.CompleteMultipartUploadOutput{
				ETag:     genETag(),
				Location: &location,
			}, nil
		},
		CreateMultipartUploadFunc: func(ctx context.Context, input *s3iot.CreateMultipartUploadInput) (*s3iot.CreateMultipartUploadOutput, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if count("create") < num["create"] {
				return nil, errTemp
			}
			if c := ch["create"]; c != nil {
				c <- input
			}
			return &s3iot.CreateMultipartUploadOutput{
				UploadID: &uploadID,
			}, nil
		},
		PutObjectFunc: func(ctx context.Context, input *s3iot.PutObjectInput) (*s3iot.PutObjectOutput, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if count("put") < num["put"] {
				input.Body.Read(make([]byte, 10))
				return nil, errTemp
			}
			if c := ch["put"]; c != nil {
				c <- input
			}
			io.Copy(buf, input.Body)
			return &s3iot.PutObjectOutput{
				ETag:     genETag(),
				Location: &location,
			}, nil
		},
		UploadPartFunc: func(ctx context.Context, input *s3iot.UploadPartInput) (*s3iot.UploadPartOutput, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if count("upload") < num["upload"] {
				input.Body.Read(make([]byte, 10))
				return nil, errTemp
			}
			if c := ch["upload"]; c != nil {
				c <- input
			}
			io.Copy(buf, input.Body)
			return &s3iot.UploadPartOutput{
				ETag: genETag(),
			}, nil
		},
	}
}

type seekErrorUploadSlicerFactory struct {
	s3iot.DefaultUploadSlicerFactory
	errs [][]error
}

func (f seekErrorUploadSlicerFactory) New(r io.Reader) (s3iot.UploadSlicer, error) {
	s, err := f.DefaultUploadSlicerFactory.New(r)
	if err != nil {
		return nil, err
	}
	return &seekErrorUploadSlicer{
		UploadSlicer: s,
		errs:         f.errs,
	}, nil
}

type seekErrorUploadSlicer struct {
	s3iot.UploadSlicer
	errs [][]error
	cnt  int
}

func (s *seekErrorUploadSlicer) NextReader() (io.ReadSeeker, func(), error) {
	r, cleanup, err := s.UploadSlicer.NextReader()
	if err != nil && err != io.EOF {
		return nil, nil, err
	}
	defer func() {
		s.cnt++
	}()
	return &iotest.SeekErrorer{ReadSeeker: r, Errs: s.errs[s.cnt]}, cleanup, err
}
