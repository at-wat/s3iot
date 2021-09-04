package s3iot

import (
	"context"
	"io"
)

type S3API interface {
	CreateMultipartUpload(ctx context.Context, input *CreateMultipartUploadInput) (*CreateMultipartUploadOutput, error)
	UploadPart(ctx context.Context, input *UploadPartInput) (*UploadPartOutput, error)
	AbortMultipartUpload(ctx context.Context, input *AbortMultipartUploadInput) (*AbortMultipartUploadOutput, error)
	CompleteMultipartUpload(ctx context.Context, input *CompleteMultipartUploadInput) (*CompleteMultipartUploadOutput, error)
	PutObject(ctx context.Context, input *PutObjectInput) (*PutObjectOutput, error)
}

type CreateMultipartUploadInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	ContentType *string
}
type CreateMultipartUploadOutput struct {
	UploadID *string
}
type UploadPartInput struct {
	Body       io.ReadSeeker
	Bucket     *string
	Key        *string
	PartNumber *int64
	UploadID   *string
}
type UploadPartOutput struct {
	ETag *string
}
type AbortMultipartUploadInput struct {
	Bucket   *string
	Key      *string
	UploadID *string
}
type AbortMultipartUploadOutput struct{}
type CompletedPart struct {
	ETag       *string
	PartNumber *int64
}
type CompleteMultipartUploadInput struct {
	Bucket         *string
	Key            *string
	CompletedParts []*CompletedPart
	UploadID       *string
}
type CompleteMultipartUploadOutput struct {
	VersionID *string
	ETag      *string
}
type PutObjectInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	Body        io.ReadSeeker
	ContentType *string
}
type PutObjectOutput struct {
	VersionID *string
	ETag      *string
}
