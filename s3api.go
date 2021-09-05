package s3iot

import (
	"context"
	"io"
)

// S3API is abstracted S3 API interface to support multiple major version of aws-sdk-go.
type S3API interface {
	CreateMultipartUpload(ctx context.Context, input *CreateMultipartUploadInput) (*CreateMultipartUploadOutput, error)
	UploadPart(ctx context.Context, input *UploadPartInput) (*UploadPartOutput, error)
	AbortMultipartUpload(ctx context.Context, input *AbortMultipartUploadInput) (*AbortMultipartUploadOutput, error)
	CompleteMultipartUpload(ctx context.Context, input *CompleteMultipartUploadInput) (*CompleteMultipartUploadOutput, error)
	PutObject(ctx context.Context, input *PutObjectInput) (*PutObjectOutput, error)
}

// CreateMultipartUploadInput represents input of CreateMultipartUpload API.
type CreateMultipartUploadInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	ContentType *string
}

// CreateMultipartUploadOutput represents output of CreateMultipartUpload API.
type CreateMultipartUploadOutput struct {
	UploadID *string
}

// UploadPartInput represents input of UploadPart API.
type UploadPartInput struct {
	Body       io.ReadSeeker
	Bucket     *string
	Key        *string
	PartNumber *int64
	UploadID   *string
}

// UploadPartOutput represents output of UploadPart API.
type UploadPartOutput struct {
	ETag *string
}

// AbortMultipartUploadInput represents input of AbortMultipartUpload API.
type AbortMultipartUploadInput struct {
	Bucket   *string
	Key      *string
	UploadID *string
}

// AbortMultipartUploadOutput represents output of AbortMultipartUpload API.
type AbortMultipartUploadOutput struct{}

// CompletedPart represents completed upload part.
type CompletedPart struct {
	ETag       *string
	PartNumber *int64
}

// CompleteMultipartUploadInput represents input of CompleteMultipartUpload API.
type CompleteMultipartUploadInput struct {
	Bucket         *string
	Key            *string
	CompletedParts []*CompletedPart
	UploadID       *string
}

// CompleteMultipartUploadOutput represents output of CompleteMultipartUpload API.
type CompleteMultipartUploadOutput struct {
	VersionID *string
	ETag      *string
}

// PutObjectInput represents input of PutObject API.
type PutObjectInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	Body        io.ReadSeeker
	ContentType *string
}

// PutObjectOutput represents output of PutObject API.
type PutObjectOutput struct {
	VersionID *string
	ETag      *string
}
