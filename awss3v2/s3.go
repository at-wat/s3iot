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

// Package awss3v2 provides s3iot.Uploader with aws-sdk-go-v2.
package awss3v2

import (
	"context"

	"github.com/at-wat/s3iot"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// NewUploader creates s3iot.Uploader from aws-sdk-go-v2 Config.
func NewUploader(c aws.Config, opts ...s3iot.UploaderOption) *s3iot.Uploader {
	u := &s3iot.Uploader{
		UpDownloaderBase: s3iot.UpDownloaderBase{
			API: NewAPI(s3.NewFromConfig(c)),
		},
	}
	for _, opt := range opts {
		opt.ApplyToUploader(u)
	}
	return u
}

// NewDownloader creates s3iot.Download from aws-sdk-go-v2 Config.
func NewDownloader(c aws.Config, opts ...s3iot.DownloaderOption) *s3iot.Downloader {
	u := &s3iot.Downloader{
		UpDownloaderBase: s3iot.UpDownloaderBase{
			API: NewAPI(s3.NewFromConfig(c)),
		},
	}
	for _, opt := range opts {
		opt.ApplyToDownloader(u)
	}
	return u
}

// NewAPI wraps s3.Client to s3iot.S3API.
func NewAPI(api S3API) s3iot.S3API {
	return &wrapper{api: api}
}

type wrapper struct {
	api S3API
}

func (w *wrapper) PutObject(ctx context.Context, input *s3iot.PutObjectInput) (*s3iot.PutObjectOutput, error) {
	var acl s3types.ObjectCannedACL
	if input.ACL != nil {
		acl = s3types.ObjectCannedACL(*input.ACL)
	}
	out, err := w.api.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      input.Bucket,
			Key:         input.Key,
			ACL:         acl,
			Body:        input.Body,
			ContentType: input.ContentType,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.PutObjectOutput{
		VersionID: out.VersionId,
		ETag:      out.ETag,
	}, nil
}

func (w *wrapper) GetObject(ctx context.Context, input *s3iot.GetObjectInput) (*s3iot.GetObjectOutput, error) {
	out, err := w.api.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:    input.Bucket,
			Key:       input.Key,
			Range:     input.Range,
			VersionId: input.VersionID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.GetObjectOutput{
		Body:          out.Body,
		ContentType:   out.ContentType,
		ContentLength: &out.ContentLength,
		ContentRange:  out.ContentRange,
		ETag:          out.ETag,
		LastModified:  out.LastModified,
		VersionID:     out.VersionId,
	}, nil
}

func (w *wrapper) CreateMultipartUpload(ctx context.Context, input *s3iot.CreateMultipartUploadInput) (*s3iot.CreateMultipartUploadOutput, error) {
	var acl s3types.ObjectCannedACL
	if input.ACL != nil {
		acl = s3types.ObjectCannedACL(*input.ACL)
	}
	out, err := w.api.CreateMultipartUpload(
		ctx,
		&s3.CreateMultipartUploadInput{
			Bucket:      input.Bucket,
			Key:         input.Key,
			ACL:         acl,
			ContentType: input.ContentType,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.CreateMultipartUploadOutput{
		UploadID: out.UploadId,
	}, nil
}

func (w *wrapper) CompleteMultipartUpload(ctx context.Context, input *s3iot.CompleteMultipartUploadInput) (*s3iot.CompleteMultipartUploadOutput, error) {
	var parts []s3types.CompletedPart
	for _, part := range input.CompletedParts {
		parts = append(parts, s3types.CompletedPart{
			ETag:       part.ETag,
			PartNumber: int32(*part.PartNumber),
		})
	}
	out, err := w.api.CompleteMultipartUpload(
		ctx,
		&s3.CompleteMultipartUploadInput{
			Bucket: input.Bucket,
			Key:    input.Key,
			MultipartUpload: &s3types.CompletedMultipartUpload{
				Parts: parts,
			},
			UploadId: input.UploadID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.CompleteMultipartUploadOutput{
		VersionID: out.VersionId,
		ETag:      out.ETag,
	}, nil
}

func (w *wrapper) AbortMultipartUpload(ctx context.Context, input *s3iot.AbortMultipartUploadInput) (*s3iot.AbortMultipartUploadOutput, error) {
	_, err := w.api.AbortMultipartUpload(
		ctx,
		&s3.AbortMultipartUploadInput{
			Bucket:   input.Bucket,
			Key:      input.Key,
			UploadId: input.UploadID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.AbortMultipartUploadOutput{}, nil
}

func (w *wrapper) UploadPart(ctx context.Context, input *s3iot.UploadPartInput) (*s3iot.UploadPartOutput, error) {
	var pn int32
	if input.PartNumber != nil {
		pn = int32(*input.PartNumber)
	}
	out, err := w.api.UploadPart(
		ctx,
		&s3.UploadPartInput{
			Body:       input.Body,
			Bucket:     input.Bucket,
			Key:        input.Key,
			PartNumber: pn,
			UploadId:   input.UploadID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.UploadPartOutput{
		ETag: out.ETag,
	}, nil

}
