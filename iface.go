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
	"io"
	"time"

	"github.com/at-wat/s3iot/contentrange"
)

// UploadSlicerFactory creates UploadSlicer for given io.Reader.
// UploadSlicer will be created for each Upload() call.
type UploadSlicerFactory interface {
	New(io.Reader) (UploadSlicer, error)
}

// UploadSlicer splits input data stream into multiple io.ReadSeekers.
type UploadSlicer interface {
	Len() int64
	NextReader() (io.ReadSeeker, func(), error)
}

// DownloadSlicerFactory creates DownloadSlicer for given io.WriterAt.
// DownloadSlicer will be created for each Download() call.
type DownloadSlicerFactory interface {
	New(io.WriterAt) (DownloadSlicer, error)
}

// DownloadSlicer splits input data stream into multiple io.WriterAt.
type DownloadSlicer interface {
	NextWriter() (io.WriterAt, contentrange.Range, error)
}

// RetryerFactory creates Retryer.
// Retryer will be created for each Upload() call.
type RetryerFactory interface {
	New(Pauser) Retryer
}

// Retryer controls upload retrying logic.
// Retryer may sleep on OnFail to control retry interval, but can be canceled by ctx.
type Retryer interface {
	OnFail(ctx context.Context, id int64, err error) bool
	OnSuccess(id int64)
}

// ErrorClassifier distinguishes given error is retryable.
type ErrorClassifier interface {
	IsRetryable(error) bool
	IsThrottle(error) (time.Duration, bool)
}

// ReadInterceptorFactory creates ReadInterceptor.
// ReadInterceptor will be created for each Upload() call.
type ReadInterceptorFactory interface {
	New() ReadInterceptor
}

// ReadInterceptor wraps io.ReadSeeker to intercept Read() calls.
type ReadInterceptor interface {
	Reader(io.ReadSeeker) io.ReadSeeker
}

// Pauser provices pause/resume interface.
type Pauser interface {
	Pause()
	Resume()
}

// DoneNotifier provices completion notifier.
type DoneNotifier interface {
	// Done returns a channel which will be closed after complete.
	Done() <-chan struct{}
}

// UploadInput represents upload destination and data.
type UploadInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	Body        io.Reader
	ContentType *string
}

// UploadOutput represents upload result.
type UploadOutput struct {
	VersionID *string
	ETag      *string
	Location  *string
}

// DownloadInput represents upload destination and data.
type DownloadInput struct {
	Bucket    *string
	Key       *string
	VersionID *string
}

// DownloadOutput represents download result.
type DownloadOutput struct {
	ContentType  *string
	ETag         *string
	LastModified *time.Time
	VersionID    *string
}

// UploadContext provides access to the upload progress and the result.
type UploadContext interface {
	// Result reutrns the upload status or error.
	Status() (UploadStatus, error)
	// Result reutrns the upload result or error.
	Result() (UploadOutput, error)

	Pauser
	DoneNotifier
}

// DownloadContext provides access to the download progress and the result.
type DownloadContext interface {
	// Result reutrns the upload status or error.
	Status() (DownloadStatus, error)
	// Result reutrns the upload result or error.
	Result() (DownloadOutput, error)

	Pauser
	DoneNotifier
}

// Status represents upload/download status.
type Status struct {
	Size          int64
	CompletedSize int64
	NumRetries    int
	Paused        bool
}

// UploadStatus represents upload status.
type UploadStatus struct {
	Status

	UploadID string
}

// DownloadStatus represents download status.
type DownloadStatus struct {
	Status
	DownloadOutput
}
