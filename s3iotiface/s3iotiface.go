package s3iotiface

import (
	"context"
	"io"

	"github.com/at-wat/s3iot"
)

// Uploader interface of s3iot.
type Uploader interface {
	Upload(ctx context.Context, input *s3iot.UploadInput) (s3iot.UploadContext, error)
}

// Downloader interface of s3iot.
type Downloader interface {
	Download(ctx context.Context, w io.WriterAt, input *s3iot.DownloadInput) (s3iot.DownloadContext, error)
}