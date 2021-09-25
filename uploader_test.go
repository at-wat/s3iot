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

				status, err := uc.Status()
				if err != nil {
					t.Fatal(err)
				}
				var expectedRetries int
				for _, num := range tt.num {
					expectedRetries += num
				}
				if status.NumRetries != expectedRetries {
					t.Errorf("Expected NumRetries: %d, got: %d", expectedRetries, status.NumRetries)
				}

				bkg, ok := uc.(s3iot.BucketKeyer)
				if !ok {
					t.Fatal("UploadContext should implement BucketKeyer")
				}
				ucBucket, ucKey := bkg.BucketKey()
				if ucBucket != bucket || ucKey != key {
					t.Errorf("Expected bucket/key: %s/%s, got: %s/%s", bucket, key, ucBucket, ucKey)
				}
			})
		}
	})
	t.Run("MultiPart", func(t *testing.T) {
		testCases := map[string]struct {
			partSize      int64
			num           map[string]int
			err           error
			calls         int
			parts         int
			readerWrapper func(io.Reader) io.Reader
		}{
			"NoAPIError": {
				partSize: 50,
				calls:    1,
				parts:    3,
			},
			"NoAPIError_SizeBoundary": {
				partSize: 64,
				calls:    1,
				parts:    2,
			},
			"NoAPIError_ReadOnly": {
				partSize: 50,
				calls:    1,
				parts:    3,
				readerWrapper: func(r io.Reader) io.Reader {
					return &iotest.ReadOnly{R: r}
				},
			},
			"NoAPIError_SizeBoundary_ReadOnly": {
				partSize: 64,
				calls:    1,
				parts:    2,
				readerWrapper: func(r io.Reader) io.Reader {
					return &iotest.ReadOnly{R: r}
				},
			},
			"NoAPIError_ReadSeekOnly": {
				partSize: 50,
				calls:    1,
				parts:    3,
				readerWrapper: func(r io.Reader) io.Reader {
					return &iotest.ReadSeekOnly{R: r.(io.ReadSeeker)}
				},
			},
			"NoAPIError_SizeBoundary_ReadSeekOnly": {
				partSize: 64,
				calls:    1,
				parts:    2,
				readerWrapper: func(r io.Reader) io.Reader {
					return &iotest.ReadSeekOnly{R: r.(io.ReadSeeker)}
				},
			},
			"OneAPIError": {
				partSize: 50,
				num:      map[string]int{"complete": 1, "create": 1, "upload": 1},
				calls:    2,
				parts:    3,
			},
			"TwoCompleteAPIError": {
				partSize: 50,
				num:      map[string]int{"complete": 2},
				err:      errTemp,
			},
			"TwoCreateAPIError": {
				partSize: 50,
				num:      map[string]int{"create": 2},
				err:      errTemp,
			},
			"TwoUploadAPIError": {
				partSize: 50,
				num:      map[string]int{"upload": 2},
				err:      errTemp,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := &bytes.Buffer{}
				api := newUploadMockAPI(buf, tt.num, nil)
				u := &s3iot.Uploader{}
				s3iot.WithAPI(api).ApplyToUploader(u)
				s3iot.WithUploadSlicer(&s3iot.DefaultUploadSlicerFactory{
					PartSize: tt.partSize,
				}).ApplyToUploader(u)
				s3iot.WithErrorClassifier(&s3iot.NaiveErrorClassifier{}).ApplyToUploader(u)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				}).ApplyToUploader(u)

				var r io.Reader = bytes.NewReader(data)
				if tt.readerWrapper != nil {
					r = tt.readerWrapper(r)
				}
				uc, err := u.Upload(context.TODO(), &s3iot.UploadInput{
					Bucket: &bucket,
					Key:    &key,
					Body:   r,
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
				if n := len(api.UploadPartCalls()); n != tt.parts+tt.num["upload"] {
					t.Fatalf("UploadPart must be called %d times, but called %d times", tt.parts+tt.num["upload"], n)
				}
				if n := len(api.CompleteMultipartUploadCalls()); n != tt.calls {
					t.Fatalf("CompleteMultipartUpload must be called %d times, but called %d times", tt.calls, n)
				}
				if len(api.PutObjectCalls()) != 0 {
					t.Fatal("PutObject must not be called")
				}

				status, err := uc.Status()
				if err != nil {
					t.Fatal(err)
				}
				var expectedRetries int
				for _, num := range tt.num {
					expectedRetries += num
				}
				if status.NumRetries != expectedRetries {
					t.Errorf("Expected NumRetries: %d, got: %d", expectedRetries, status.NumRetries)
				}

				expectedETag := fmt.Sprintf("TAG%d", tt.parts+1)
				if *out.ETag != expectedETag {
					t.Errorf("Expected ETag: %s, got: %s", expectedETag, *out.ETag)
				}
				if *out.Location != "s3://url" {
					t.Errorf("Expected Location: s3://url, got: %s", *out.Location)
				}
				if !bytes.Equal(data, buf.Bytes()) {
					t.Error("Uploaded data differs")
				}

				comp := api.CompleteMultipartUploadCalls()
				if n := len(comp[0].Input.CompletedParts); n != tt.parts {
					t.Fatalf("Expected %d parts, actually %d parts", tt.parts, n)
				}
				for i, expected := range []string{"TAG1", "TAG2", "TAG3"} {
					if i >= tt.parts {
						break
					}
					tag := comp[0].Input.CompletedParts[i].ETag
					if expected != *tag {
						t.Errorf("Part %d must have ETag: %s, actual: %s", i, expected, *tag)
					}
				}

				bkg, ok := uc.(s3iot.BucketKeyer)
				if !ok {
					t.Fatal("UploadContext should implement BucketKeyer")
				}
				ucBucket, ucKey := bkg.BucketKey()
				if ucBucket != bucket || ucKey != key {
					t.Errorf("Expected bucket/key: %s/%s, got: %s/%s", bucket, key, ucBucket, ucKey)
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
		if !status.Paused {
			t.Error("Paused flag must be set")
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

		status, err = uc.Status()
		if err != nil {
			t.Fatal(err)
		}
		if status.Paused {
			t.Error("Paused flag must not be set")
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
			partSize                   int64
			innerSeekerFailAtSeekerMax int
			innerSeekerFailAtMax       int
		}{
			"Single": {
				partSize:                   150,
				innerSeekerFailAtSeekerMax: 1,
				innerSeekerFailAtMax:       1,
			},
			"SingleBoundary": {
				partSize:                   128,
				innerSeekerFailAtSeekerMax: 1,
				innerSeekerFailAtMax:       1,
			},
			"Multi": {
				partSize:                   50,
				innerSeekerFailAtSeekerMax: 3,
				innerSeekerFailAtMax:       2,
			},
			"MultiBoundary": {
				partSize:                   64,
				innerSeekerFailAtSeekerMax: 2,
				innerSeekerFailAtMax:       2,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				t.Run("OuterSeeker", func(t *testing.T) {
					for n := 0; n < 2; n++ {
						errs := []error{errSeekFailure}
						for i := 0; i < n; i++ {
							errs = append([]error{nil}, errs...)
						}

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
					}
				})
				t.Run("InnerSeeker", func(t *testing.T) {
					for n := 0; n < tt.innerSeekerFailAtMax; n++ {
						for m := 0; m < tt.innerSeekerFailAtSeekerMax; m++ {
							errs := [][]error{{errSeekFailure}}
							for i := 0; i < n; i++ {
								errs[0] = append([]error{nil}, errs[0]...)
							}
							for i := 0; i < m; i++ {
								errs = append([][]error{{nil}}, errs...)
							}

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

							t.Run(fmt.Sprintf("SeekErrorAt%d_%d", m, n), func(t *testing.T) {
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
						}
					}
				})
			})
		}
	})
	t.Run("SlicerError", func(t *testing.T) {
		errSliceFailure := errors.New("slice error")

		testCases := map[string]struct {
			partSize int64
			errAt    int
			err0     error
			err1     error
		}{
			"Single": {
				partSize: 150,
				errAt:    0,
				err0:     errSliceFailure,
			},
			"Multi": {
				partSize: 50,
				errAt:    0,
				err0:     errSliceFailure,
			},
			"Multi2": {
				partSize: 64,
				errAt:    1,
				err1:     errSliceFailure,
			},
		}

		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				u := &s3iot.Uploader{}
				s3iot.WithRetryer(&s3iot.NoRetryerFactory{}).ApplyToUploader(u)
				s3iot.WithAPI(newUploadMockAPI(&bytes.Buffer{}, nil, nil)).ApplyToUploader(u)
				s3iot.WithUploadSlicer(
					&nextErrorUploadSlicerFactory{
						DefaultUploadSlicerFactory: s3iot.DefaultUploadSlicerFactory{
							PartSize: tt.partSize,
						},
						err:   errSliceFailure,
						errAt: tt.errAt,
					},
				).ApplyToUploader(u)

				uc, err := u.Upload(context.TODO(), &s3iot.UploadInput{
					Bucket: &bucket,
					Key:    &key,
					Body:   bytes.NewReader(data),
				})
				if !errors.Is(err, tt.err0) {
					t.Fatalf("Expected error: '%v', got: '%v'", tt.err0, err)
				}
				if err != nil {
					return
				}

				select {
				case <-time.After(time.Second):
					t.Fatal("Timeout")
				case <-uc.Done():
				}

				if _, err = uc.Result(); !errors.Is(err, tt.err1) {
					t.Fatalf("Expected error: '%v', got: '%v'", tt.err1, err)
				}
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

type nextErrorUploadSlicerFactory struct {
	s3iot.DefaultUploadSlicerFactory
	err   error
	errAt int
}

func (f nextErrorUploadSlicerFactory) New(r io.Reader) (s3iot.UploadSlicer, error) {
	s, err := f.DefaultUploadSlicerFactory.New(r)
	if err != nil {
		return nil, err
	}
	return &nextErrorUploadSlicer{
		UploadSlicer: s,
		err:          f.err,
		errAt:        f.errAt,
	}, nil
}

type nextErrorUploadSlicer struct {
	s3iot.UploadSlicer
	err   error
	errAt int
	cnt   int
}

func (s *nextErrorUploadSlicer) NextReader() (io.ReadSeeker, func(), error) {
	if s.cnt >= s.errAt {
		return nil, nil, s.err
	}
	s.cnt++
	return s.UploadSlicer.NextReader()
}
