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
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/at-wat/s3iot/contentrange"
)

// Download errors.
var (
	ErrChangedDuringDownload    = errors.New("object is changed during download")
	ErrUnexpectedServerResponse = errors.New("unexpected server response")
)

// Download a file to S3.
func (u Downloader) Download(ctx context.Context, w io.WriterAt, input *DownloadInput) (DownloadContext, error) {
	if u.DownloadSlicerFactory == nil {
		u.DownloadSlicerFactory = &DefaultDownloadSlicerFactory{}
	}
	if u.RetryerFactory == nil {
		u.RetryerFactory = DefaultRetryer
	}
	if u.ErrorClassifier == nil {
		u.ErrorClassifier = DefaultErrorClassifier
	}
	slicer, err := u.DownloadSlicerFactory.New(w)
	if err != nil {
		return nil, err
	}
	dc := &downloadContext{
		api:           u.API,
		slicer:        slicer,
		errClassifier: u.ErrorClassifier,
		input:         input,
		done:          make(chan struct{}),
		paused:        make(chan struct{}),
	}
	dc.retryer = u.RetryerFactory.New(dc)
	close(dc.paused)
	dc.resumeOnce.Do(func() {})
	go dc.multi(ctx)
	return dc, nil
}

type downloadContext struct {
	api           S3API
	slicer        DownloadSlicer
	retryer       Retryer
	errClassifier ErrorClassifier
	input         *DownloadInput

	status DownloadStatus
	output DownloadOutput
	err    error

	paused     chan struct{}
	resumeOnce sync.Once

	mu   sync.RWMutex
	done chan struct{}
}

func (dc *downloadContext) Done() <-chan struct{} {
	return dc.done
}

func (dc *downloadContext) Pause() {
	dc.mu.Lock()
	dc.paused = make(chan struct{})
	dc.resumeOnce = sync.Once{}
	dc.status.Paused = true
	dc.mu.Unlock()
}

func (dc *downloadContext) Resume() {
	dc.mu.Lock()
	dc.resumeOnce.Do(func() {
		close(dc.paused)
	})
	dc.status.Paused = false
	dc.mu.Unlock()
}

func (dc *downloadContext) Status() (DownloadStatus, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.status, dc.err
}

func (dc *downloadContext) Result() (DownloadOutput, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.output, dc.err
}

func (dc *downloadContext) countRetry() {
	dc.mu.Lock()
	dc.status.NumRetries++
	dc.mu.Unlock()
}

func (dc *downloadContext) pauseCheck(ctx context.Context) {
	dc.mu.RLock()
	paused := dc.paused
	dc.mu.RUnlock()

	select {
	case <-paused:
	case <-ctx.Done():
	}
}

func (dc *downloadContext) multi(ctx context.Context) {
	for i := int64(1); ; i++ {
		i := i
		w, rn, err := dc.slicer.NextWriter()
		if err != nil {
			dc.fail(err)
			return
		}
		var n int64
		var fatal bool
		if err := withRetry(ctx, i, dc.retryer, dc.errClassifier, func() error {
			dc.pauseCheck(ctx)
			r := rn.String()
			out, err := dc.api.GetObject(ctx, &GetObjectInput{
				Bucket:    dc.input.Bucket,
				Key:       dc.input.Key,
				Range:     &r,
				VersionID: dc.input.VersionID,
			})
			if err != nil {
				dc.countRetry()
				return err
			}
			defer out.Body.Close()

			rn2, err := contentrange.ParseContentRange(*out.ContentRange)
			if err != nil {
				dc.countRetry()
				return err
			}
			if rn.Start != rn2.Start {
				dc.countRetry()
				return fmt.Errorf(
					"requested range=%s, returned range=%s: %w",
					rn, rn2,
					ErrUnexpectedServerResponse,
				)
			}
			rn = *rn2

			dc.mu.Lock()
			if dc.status.ETag != nil && *dc.status.ETag != *out.ETag {
				// File is changed during download.
				err := fmt.Errorf(
					"initial ETag=%s, current ETag=%s: %w",
					*dc.status.ETag, *out.ETag,
					ErrChangedDuringDownload,
				)
				dc.mu.Unlock()
				dc.fail(err)
				fatal = true
				return nil
			}
			dc.status.Size = rn2.Size
			dc.status.ContentType = out.ContentType
			dc.status.ETag = out.ETag
			dc.status.LastModified = out.LastModified
			dc.status.VersionID = out.VersionID
			dc.mu.Unlock()

			n, err = io.Copy(&atWriter{w: w}, out.Body)
			if err != nil {
				dc.countRetry()
				return err
			}
			return nil
		}); err != nil {
			dc.fail(err)
			return
		}
		if fatal {
			return
		}

		dc.mu.Lock()
		dc.status.CompletedSize += n
		done := dc.status.CompletedSize >= dc.status.Size
		dc.mu.Unlock()

		if done {
			dc.success(dc.status.DownloadOutput)
			return
		}
	}
}

func (dc *downloadContext) fail(err error) {
	dc.mu.Lock()
	dc.err = err
	dc.mu.Unlock()
	close(dc.done)
}

func (dc *downloadContext) success(out DownloadOutput) {
	dc.mu.Lock()
	dc.output = out
	dc.mu.Unlock()
	close(dc.done)
}
