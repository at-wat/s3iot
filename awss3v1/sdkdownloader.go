package awss3v1

import (
	"context"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/s3iotiface"
)

// NewAWSSDKDownloader wraps aws-sdk-go s3manager as s3iotiface.Downloader.
// Some fields of download status and result is not provided.
// Pause/resume feature is unavailable.
func NewAWSSDKDownloader(u s3manageriface.DownloaderAPI) s3iotiface.Downloader {
	return &sdkDownloader{u: u}
}

type sdkDownloader struct {
	u s3manageriface.DownloaderAPI
}

type sdkDownloaderContext struct {
	s3iot.DoneNotifier

	mu sync.RWMutex

	err    error
	output s3iot.DownloadOutput
	status s3iot.DownloadStatus
}

func (u *sdkDownloader) Download(ctx context.Context, w io.WriterAt, input *s3iot.DownloadInput) (s3iot.DownloadContext, error) {
	doneCtx, cancel := context.WithCancel(context.Background())
	dc := &sdkDownloaderContext{
		DoneNotifier: doneCtx,
		output: s3iot.DownloadOutput{
			VersionID: input.VersionID,
		},
	}
	go func() {
		n, err := u.u.DownloadWithContext(ctx, w, &s3.GetObjectInput{
			Bucket:    input.Bucket,
			Key:       input.Key,
			VersionId: input.VersionID,
		})
		dc.mu.Lock()
		dc.err = err
		dc.status.CompletedSize = n
		dc.mu.Unlock()
		cancel()
	}()
	return dc, nil
}

func (c *sdkDownloaderContext) Pause()  {}
func (c *sdkDownloaderContext) Resume() {}

func (c *sdkDownloaderContext) Status() (s3iot.DownloadStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status, c.err
}

func (c *sdkDownloaderContext) Result() (s3iot.DownloadOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.output, c.err
}
