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
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/contentrange"
	"github.com/at-wat/s3iot/internal/iotest"
	mock_s3iot "github.com/at-wat/s3iot/internal/moq/s3iot"
)

func TestDownloader(t *testing.T) {
	var (
		bucket = "Bucket"
		key    = "Key"
	)
	data := make([]byte, 128)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	t.Run("MultiPart", func(t *testing.T) {
		testCases := map[string]struct {
			num   int
			err   error
			calls int
		}{
			"NoAPIError": {
				calls: 3,
			},
			"OneAPIError": {
				num:   1,
				calls: 4,
			},
			"TwoAPIErrors": {
				num: 2,
				err: errTemp,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := iotest.BufferAt(make([]byte, 128))
				api := newDownloadMockAPI(t, data, tt.num, nil, nil)
				d := &s3iot.Downloader{}
				s3iot.WithAPI(api).ApplyToDownloader(d)
				s3iot.WithDownloadSlicer(
					&s3iot.DefaultDownloadSlicerFactory{PartSize: 50},
				).ApplyToDownloader(d)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				}).ApplyToDownloader(d)

				dc, err := d.Download(context.TODO(), buf, &s3iot.DownloadInput{
					Bucket: &bucket,
					Key:    &key,
				})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-time.After(time.Second):
					t.Fatal("Timeout")
				case <-dc.Done():
				}
				out, err := dc.Result()
				if tt.err == nil {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					if !errors.Is(err, errTemp) {
						t.Fatalf("Expected error: '%v', got: '%v'", errTemp, err)
					}
					return
				}

				if n := len(api.GetObjectCalls()); n != tt.calls {
					t.Fatalf("GetObject must be called %d times, but called %d times", tt.calls, n)
				}

				status, err := dc.Status()
				if err != nil {
					t.Fatal(err)
				}
				if status.NumRetries != tt.num {
					t.Errorf("Expected NumRetries: %d, got: %d", tt.num, status.NumRetries)
				}

				if *out.ETag != "TAG0" {
					t.Errorf("Expected ETag: TAG0, got: %s", *out.ETag)
				}
				if !bytes.Equal(data, []byte(buf)) {
					t.Error("Downloaded data differs")
				}

				bkg, ok := dc.(s3iot.BucketKeyer)
				if !ok {
					t.Fatal("UploadContext should implement BucketKeyer")
				}
				dcBucket, dcKey := bkg.BucketKey()
				if dcBucket != bucket || dcKey != key {
					t.Errorf("Expected bucket/key: %s/%s, got: %s/%s", bucket, key, dcBucket, dcKey)
				}
			})
		}
	})
	t.Run("RetryOnRangeError", func(t *testing.T) {
		testCases := map[string]struct {
			contentRange string
			err          error
		}{
			"InvalidRange": {
				contentRange: "dummy",
				err:          contentrange.ErrInvalidFormat,
			},
			"WrongRange": {
				contentRange: "bytes 1-2/100",
				err:          s3iot.ErrUnexpectedServerResponse,
			},
		}
		for name, tt := range testCases {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := iotest.BufferAt(make([]byte, 128))
				api := newDownloadMockAPI(t, data, 0, nil, nil)
				getObj := api.GetObjectFunc
				api.GetObjectFunc = func(ctx context.Context, input *s3iot.GetObjectInput) (*s3iot.GetObjectOutput, error) {
					out, err := getObj(ctx, input)
					out.ContentRange = &tt.contentRange
					return out, err
				}
				d := &s3iot.Downloader{}
				s3iot.WithAPI(api).ApplyToDownloader(d)
				s3iot.WithDownloadSlicer(
					&s3iot.DefaultDownloadSlicerFactory{PartSize: 50},
				).ApplyToDownloader(d)
				s3iot.WithRetryer(&s3iot.ExponentialBackoffRetryerFactory{
					WaitBase: time.Millisecond,
					RetryMax: 1,
				}).ApplyToDownloader(d)

				dc, err := d.Download(context.TODO(), buf, &s3iot.DownloadInput{
					Bucket: &bucket,
					Key:    &key,
				})
				if err != nil {
					t.Fatal(err)
				}
				select {
				case <-time.After(time.Second):
					t.Fatal("Timeout")
				case <-dc.Done():
				}
				if _, err := dc.Result(); !errors.Is(err, tt.err) {
					t.Fatalf("Expected error: '%v', got: '%v'", tt.err, err)
				}

				if n := len(api.GetObjectCalls()); n != 2 {
					t.Fatalf("GetObject must be called twice, but called %d times", n)
				}
			})
		}
	})
	t.Run("WriteError", func(t *testing.T) {
		d := &s3iot.Downloader{}
		s3iot.WithAPI(
			newDownloadMockAPI(t, data, 0, nil, nil),
		).ApplyToDownloader(d)

		errWrite := errors.New("write error")
		buf := iotest.WriteAtErrorer{Err: errWrite}

		dc, err := d.Download(context.TODO(), buf, &s3iot.DownloadInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-dc.Done():
		}
		if _, err := dc.Result(); !errors.Is(err, errWrite) {
			t.Fatalf("Expected error: '%v', got: '%v'", errWrite, err)
		}
	})

	t.Run("PauseResume", func(t *testing.T) {
		for name, tt := range map[string]struct {
			forcePause    bool
			expectedCalls int
		}{
			"NoForcePause": {
				forcePause:    false,
				expectedCalls: 3,
			},
			"ForcePause": {
				forcePause:    true,
				expectedCalls: 4,
			},
		} {
			tt := tt
			t.Run(name, func(t *testing.T) {
				buf := iotest.BufferAt(make([]byte, 128))
				chDownload := make(chan interface{})
				api := newDownloadMockAPI(t, data, 0, chDownload, nil)
				d := &s3iot.Downloader{
					UpDownloaderBase: s3iot.UpDownloaderBase{
						ForcePause: tt.forcePause,
					},
				}
				s3iot.WithAPI(api).ApplyToDownloader(d)
				s3iot.WithDownloadSlicer(
					&s3iot.DefaultDownloadSlicerFactory{PartSize: 50},
				).ApplyToDownloader(d)
				s3iot.WithRetryer(nil).ApplyToDownloader(d)

				uc, err := d.Download(context.TODO(), buf, &s3iot.DownloadInput{
					Bucket: &bucket,
					Key:    &key,
				})
				if err != nil {
					t.Fatal(err)
				}

				select {
				case <-time.After(time.Second):
					t.Fatal("Timeout")
				case <-chDownload:
				}

				uc.Pause()
				go func() {
					<-chDownload
					<-chDownload
				}()

				time.Sleep(50 * time.Millisecond)
				status, err := uc.Status()
				if err != nil {
					t.Fatal(err)
				}
				if !status.Paused {
					t.Error("Paused flag must be set")
				}
				if *status.ETag != "TAG0" {
					t.Errorf("Expected ETag: TAG0, got: %s", *status.ETag)
				}

				select {
				case <-time.After(500 * time.Millisecond):
				case <-uc.Done():
					t.Fatal("Download should be paused")
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

				if n := len(api.GetObjectCalls()); n != tt.expectedCalls {
					t.Fatalf("GetObject must be called %d times, but called %d times", tt.expectedCalls, n)
				}

				if !bytes.Equal(data, []byte(buf)) {
					t.Error("Downloaded data differs")
				}
			})
		}
	})
	t.Run("CancelDuringPause", func(t *testing.T) {
		buf := iotest.BufferAt(make([]byte, 128))
		chDownload := make(chan interface{})
		api := newDownloadMockAPI(t, data, 0, chDownload, nil)
		d := &s3iot.Downloader{}
		s3iot.WithAPI(api).ApplyToDownloader(d)
		s3iot.WithDownloadSlicer(
			&s3iot.DefaultDownloadSlicerFactory{PartSize: 50},
		).ApplyToDownloader(d)
		s3iot.WithRetryer(nil).ApplyToDownloader(d)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		uc, err := d.Download(ctx, buf, &s3iot.DownloadInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(50 * time.Millisecond)
		uc.Pause()

		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-chDownload:
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
	t.Run("FileChangedDuringDownload", func(t *testing.T) {
		buf := iotest.BufferAt(make([]byte, 128))
		api := newDownloadMockAPI(t, data, 0, nil, []string{"TAG0", "TAG1"})
		d := &s3iot.Downloader{}
		s3iot.WithAPI(api).ApplyToDownloader(d)
		s3iot.WithDownloadSlicer(
			&s3iot.DefaultDownloadSlicerFactory{PartSize: 50},
		).ApplyToDownloader(d)
		s3iot.WithRetryer(nil).ApplyToDownloader(d)

		uc, err := d.Download(context.TODO(), buf, &s3iot.DownloadInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-uc.Done():
		}
		if _, err = uc.Result(); !errors.Is(err, s3iot.ErrChangedDuringDownload) {
			t.Fatalf("Expected error: '%v', got: '%v'", s3iot.ErrChangedDuringDownload, err)
		}
	})
	t.Run("DefaultSlicer", func(t *testing.T) {
		buf := iotest.BufferAt(make([]byte, 128))
		api := newDownloadMockAPI(t, data, 0, nil, nil)
		d := &s3iot.Downloader{}
		s3iot.WithAPI(api).ApplyToDownloader(d)

		dc, err := d.Download(context.TODO(), buf, &s3iot.DownloadInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-dc.Done():
		}
		if _, err := dc.Result(); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, []byte(buf)) {
			t.Error("Downloaded data differs")
		}
	})
	t.Run("ContextCanceled", func(t *testing.T) {
		buf := iotest.BufferAt(make([]byte, 128))
		api := newDownloadMockAPI(t, data, 0, nil, nil)
		d := &s3iot.Downloader{}
		s3iot.WithAPI(api).ApplyToDownloader(d)
		s3iot.WithDownloadSlicer(&s3iot.DefaultDownloadSlicerFactory{PartSize: 50}).ApplyToDownloader(d)
		s3iot.WithRetryer(nil).ApplyToDownloader(d)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		uc, err := d.Download(ctx, buf, &s3iot.DownloadInput{
			Bucket: &bucket,
			Key:    &key,
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

func newDownloadMockAPI(t *testing.T, data []byte, num int, ch chan interface{}, etags []string) *mock_s3iot.MockS3API {
	var mu sync.Mutex
	var cnt int
	count := func() int {
		mu.Lock()
		count := cnt
		cnt++
		mu.Unlock()
		return count
	}

	return &mock_s3iot.MockS3API{
		GetObjectFunc: func(ctx context.Context, input *s3iot.GetObjectInput) (*s3iot.GetObjectOutput, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			i := count()
			if i < num {
				return nil, errTemp
			}
			etag := "TAG0"
			if len(etags) > i {
				etag = etags[i]
			}
			if ch != nil {
				select {
				case ch <- input:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			r, err := contentrange.Parse(*input.Range)
			if err != nil {
				t.Error(err)
			}
			r.Size = int64(len(data))
			if r.End > int64(len(data)) {
				r.End = int64(len(data)) - 1
			}
			cr := r.ContentRange()
			return &s3iot.GetObjectOutput{
				Body:         io.NopCloser(bytes.NewReader(data[r.Start : r.End+1])),
				ContentRange: &cr,
				ETag:         &etag,
			}, nil
		},
	}
}
