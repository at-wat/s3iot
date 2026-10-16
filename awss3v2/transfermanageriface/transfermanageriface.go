package transfermanageriface

import (
	"context"

	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
)

// Client is an interface of github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager.Client.
type Client interface {
	DownloadObject(ctx context.Context, input *tm.DownloadObjectInput, opts ...func(*tm.Options)) (*tm.DownloadObjectOutput, error)
	UploadObject(ctx context.Context, input *tm.UploadObjectInput, opts ...func(*tm.Options)) (*tm.UploadObjectOutput, error)
}

var _ Client = &tm.Client{}
