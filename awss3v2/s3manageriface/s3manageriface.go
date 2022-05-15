package s3manageriface

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Uploader is an interface of github.com/aws/aws-sdk-go-v2/feature/s3/manager.Uploader.
type Uploader interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

// Downloader is an interface of github.com/aws/aws-sdk-go-v2/feature/s3/manager.Downloader.
type Downloader interface {
	Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*manager.Downloader)) (n int64, err error)
}
