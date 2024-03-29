package awss3v2

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/awss3v2-1.22.2/s3manageriface"
	"github.com/at-wat/s3iot/s3iotiface"
)

// NewAWSSDKUploader wraps aws-sdk-go-v2 s3manager as s3iotiface.Uploader.
// Some fields of download status and result is not provided.
// Pause/resume feature is unavailable.
func NewAWSSDKUploader(u s3manageriface.Uploader) s3iotiface.Uploader {
	return &sdkUploader{u: u}
}

type sdkUploader struct {
	u s3manageriface.Uploader
}

type sdkUploaderContext struct {
	s3iot.DoneNotifier

	mu sync.RWMutex

	err    error
	output s3iot.UploadOutput
}

func (u *sdkUploader) Upload(ctx context.Context, input *s3iot.UploadInput) (s3iot.UploadContext, error) {
	doneCtx, cancel := context.WithCancel(context.Background())
	uc := &sdkUploaderContext{DoneNotifier: doneCtx}
	var acl types.ObjectCannedACL
	if input.ACL != nil {
		acl = types.ObjectCannedACL(*input.ACL)
	}
	in := &s3.PutObjectInput{
		ACL:         acl,
		Body:        input.Body,
		Bucket:      input.Bucket,
		ContentType: input.ContentType,
		Key:         input.Key,
	}
	go func() {
		out, err := u.u.Upload(ctx, in)
		uc.mu.Lock()
		uc.err = err
		if out != nil {
			uc.output.ETag = out.ETag
			uc.output.VersionID = out.VersionID
			uc.output.Location = &out.Location
		}
		uc.mu.Unlock()
		cancel()
	}()
	return uc, nil
}

func (c *sdkUploaderContext) Pause()  {}
func (c *sdkUploaderContext) Resume() {}

func (c *sdkUploaderContext) Status() (s3iot.UploadStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return s3iot.UploadStatus{}, c.err
}

func (c *sdkUploaderContext) Result() (s3iot.UploadOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.output, c.err
}
