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

// Package s3api provides abstracted S3 API interface to support multiple major version of aws-sdk-go.
package s3api

import (
	"context"
	"io"
	"time"
)

// UploadAPI interface.
type UploadAPI interface {
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
	Location  *string
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
	Location  *string
}

// DownloadAPI interface.
type DownloadAPI interface {
	GetObject(ctx context.Context, input *GetObjectInput) (*GetObjectOutput, error)
}

// GetObjectInput represents input of GetObject API.
type GetObjectInput struct {
	Bucket    *string
	Key       *string
	Range     *string
	VersionID *string
}

// GetObjectOutput represents output of GetObject API.
type GetObjectOutput struct {
	Body          io.ReadCloser
	ContentType   *string
	ContentLength *int64
	ContentRange  *string
	ETag          *string
	LastModified  *time.Time
	VersionID     *string
}

// DeleteAPI interface.
type DeleteAPI interface {
	DeleteObject(ctx context.Context, input *DeleteObjectInput) (*DeleteObjectOutput, error)
}

// DeleteObjectInput represents input of Delete API.
type DeleteObjectInput struct {
	Bucket    *string
	Key       *string
	VersionID *string
}

// DeleteObjectOutput represents output of Delete API.
type DeleteObjectOutput struct {
	VersionID *string
}

// ListAPI interface.
type ListAPI interface {
	ListObjectsV2(ctx context.Context, input *ListObjectsV2Input) (*ListObjectsV2Output, error)
}

// ListObjectsV2Input represents input of List API.
type ListObjectsV2Input struct {
	Bucket            *string
	ContinuationToken *string
	MaxKeys           int
	Prefix            *string
}

// ListObjectsV2Output represents output of List API.
type ListObjectsV2Output struct {
	Contents              []Object
	KeyCount              int
	NextContinuationToken *string
}

// Object represents S3 object.
type Object struct {
	ETag         *string
	Key          *string
	LastModified *time.Time
	Size         int64
}

// UpDownloadAPI is the interface that groups Upload and Download APIs.
type UpDownloadAPI interface {
	UploadAPI
	DownloadAPI
}

// S3API is the interface that groups all S3 APIs.
type S3API interface {
	UpDownloadAPI
	DeleteAPI
	ListAPI
}
