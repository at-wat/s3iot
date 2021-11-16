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
	"sync"
)

// UpDownloaderBase stores downloader/uploader base objects.
type UpDownloaderBase struct {
	API             S3API
	RetryerFactory  RetryerFactory
	ErrorClassifier ErrorClassifier
	ForcePause      bool
}

// Uploader implements S3 uploader with configurable retry and bandwidth limit.
type Uploader struct {
	UpDownloaderBase

	UploadSlicerFactory    UploadSlicerFactory
	ReadInterceptorFactory ReadInterceptorFactory
}

// Downloader implements S3 downloader with configurable retry and bandwidth limit.
type Downloader struct {
	UpDownloaderBase

	DownloadSlicerFactory DownloadSlicerFactory
}

// UploaderOption sets optional parameter to the Uploader.
type UploaderOption interface {
	ApplyToUploader(*Uploader)
}

// DownloaderOption sets optional parameter to the Downloader.
type DownloaderOption interface {
	ApplyToDownloader(*Downloader)
}

// UpDownloaderOption sets optional parameter to the Uploader or Downloader.
type UpDownloaderOption interface {
	UploaderOption
	DownloaderOption
}

// UploaderOptionFn is functional option for Uploader.
type UploaderOptionFn func(*Uploader)

// ApplyToUploader apply the option to the Uploader.
func (f UploaderOptionFn) ApplyToUploader(u *Uploader) {
	f(u)
}

// DownloaderOptionFn is functional option for Downloader.
type DownloaderOptionFn func(*Downloader)

// ApplyToDownloader apply the option to the Downloader.
func (f DownloaderOptionFn) ApplyToDownloader(d *Downloader) {
	f(d)
}

// UpDownloaderOptionFn is functional option for Uploader/Downloader.
type UpDownloaderOptionFn func(*UpDownloaderBase)

// ApplyToUploader apply the option to the Uploader.
func (f UpDownloaderOptionFn) ApplyToUploader(u *Uploader) {
	f(&u.UpDownloaderBase)
}

// ApplyToDownloader apply the option to the Downloader.
func (f UpDownloaderOptionFn) ApplyToDownloader(d *Downloader) {
	f(&d.UpDownloaderBase)
}

// WithAPI sets S3 API.
func WithAPI(a S3API) UpDownloaderOption {
	return UpDownloaderOptionFn(func(u *UpDownloaderBase) {
		u.API = a
	})
}

// WithRetryer sets RetryerFactor.
func WithRetryer(r RetryerFactory) UpDownloaderOption {
	return UpDownloaderOptionFn(func(u *UpDownloaderBase) {
		u.RetryerFactory = r
	})
}

// WithErrorClassifier sets ErrorClassifierFactor.
func WithErrorClassifier(ec ErrorClassifier) UpDownloaderOption {
	return UpDownloaderOptionFn(func(u *UpDownloaderBase) {
		u.ErrorClassifier = ec
	})
}

// WithUploadSlicer sets UploadSlicerFactory to Uploader.
func WithUploadSlicer(s UploadSlicerFactory) UploaderOption {
	return UploaderOptionFn(func(u *Uploader) {
		u.UploadSlicerFactory = s
	})
}

// WithReadInterceptor sets ReadInterceptorFactory to Uploader.
func WithReadInterceptor(i ReadInterceptorFactory) UploaderOption {
	return UploaderOptionFn(func(u *Uploader) {
		u.ReadInterceptorFactory = i
	})
}

// WithDownloadSlicer sets DownloadSlicerFactory to Downloader.
func WithDownloadSlicer(s DownloadSlicerFactory) DownloaderOption {
	return DownloaderOptionFn(func(u *Downloader) {
		u.DownloadSlicerFactory = s
	})
}

type upDownloadContext struct {
	api           S3API
	retryer       Retryer
	errClassifier ErrorClassifier
	forcePause    bool

	err error

	paused     chan struct{}
	resumeOnce sync.Once

	mu   sync.RWMutex
	done chan struct{}

	statusPaused      *bool
	statusNumRetries  *int
	cancelCurrentCall func()
}

func newUpDownloadContext(api S3API, retryerFactory RetryerFactory, errClassifier ErrorClassifier, forcePause bool) *upDownloadContext {
	c := &upDownloadContext{
		api:           api,
		errClassifier: errClassifier,
		done:          make(chan struct{}),
		paused:        make(chan struct{}),
		forcePause:    forcePause,
	}
	c.retryer = retryerFactory.New(c)
	close(c.paused)
	c.resumeOnce.Do(func() {})
	return c
}

func (c *upDownloadContext) setStatePtr(paused *bool, numRetries *int) {
	c.statusPaused = paused
	c.statusNumRetries = numRetries
}

func (c *upDownloadContext) Done() <-chan struct{} {
	return c.done
}
func (c *upDownloadContext) Pause() {
	c.mu.Lock()
	c.paused = make(chan struct{})
	c.resumeOnce = sync.Once{}
	*c.statusPaused = true
	if c.cancelCurrentCall != nil && c.forcePause {
		c.cancelCurrentCall()
	}
	c.mu.Unlock()
}

func (c *upDownloadContext) Resume() {
	c.mu.Lock()
	c.resumeOnce.Do(func() {
		close(c.paused)
	})
	*c.statusPaused = false
	c.mu.Unlock()
}

func (c *upDownloadContext) pauseCheck(ctx context.Context) {
	c.mu.RLock()
	paused := c.paused
	c.mu.RUnlock()

	select {
	case <-paused:
	case <-ctx.Done():
	}
}

func (c *upDownloadContext) currentCallContext(ctx context.Context) (context.Context, func()) {
	ctx2, cancel := context.WithCancel(ctx)
	c.mu.RLock()
	c.cancelCurrentCall = cancel
	c.mu.RUnlock()
	return ctx2, cancel
}

func (c *upDownloadContext) countRetry() {
	c.mu.Lock()
	*c.statusNumRetries++
	c.mu.Unlock()
}
