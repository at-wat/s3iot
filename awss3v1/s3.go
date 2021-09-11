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

// Package awss3v1 provides s3iot.Uploader with aws-sdk-go (v1).
package awss3v1

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/at-wat/s3iot"
)

// NewUploader creates s3iot.Uploader from aws-sdk-go ConfigProvider (like Session).
func NewUploader(c client.ConfigProvider, opts ...s3iot.UploaderOption) *s3iot.Uploader {
	u := &s3iot.Uploader{
		API: NewAPI(s3.New(c)),
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// NewAPI wraps s3iface.S3API to s3iot.S3API.
func NewAPI(api s3iface.S3API) s3iot.S3API {
	return &wrapper{api: api}
}

type wrapper struct {
	api s3iface.S3API
}

func (w *wrapper) PutObject(ctx context.Context, input *s3iot.PutObjectInput) (*s3iot.PutObjectOutput, error) {
	out, err := w.api.PutObjectWithContext(
		aws.Context(ctx),
		&s3.PutObjectInput{
			Bucket:      input.Bucket,
			Key:         input.Key,
			ACL:         input.ACL,
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
	out, err := w.api.GetObjectWithContext(
		aws.Context(ctx),
		&s3.GetObjectInput{
			Bucket:     input.Bucket,
			Key:        input.Key,
			PartNumber: input.PartNumber,
			VersionId:  input.VersionID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.GetObjectOutput{
		Body:         out.Body,
		ContentType:  out.ContentType,
		ETag:         out.ETag,
		LastModified: out.LastModified,
		PartsCount:   out.PartsCount,
		VersionID:    out.VersionId,
	}, nil
}

func (w *wrapper) CreateMultipartUpload(ctx context.Context, input *s3iot.CreateMultipartUploadInput) (*s3iot.CreateMultipartUploadOutput, error) {
	out, err := w.api.CreateMultipartUploadWithContext(
		aws.Context(ctx),
		&s3.CreateMultipartUploadInput{
			Bucket:      input.Bucket,
			Key:         input.Key,
			ACL:         input.ACL,
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
	var parts []*s3.CompletedPart
	for _, part := range input.CompletedParts {
		parts = append(parts, &s3.CompletedPart{
			ETag:       part.ETag,
			PartNumber: part.PartNumber,
		})
	}
	out, err := w.api.CompleteMultipartUploadWithContext(
		aws.Context(ctx),
		&s3.CompleteMultipartUploadInput{
			Bucket: input.Bucket,
			Key:    input.Key,
			MultipartUpload: &s3.CompletedMultipartUpload{
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
	_, err := w.api.AbortMultipartUploadWithContext(
		aws.Context(ctx),
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
	out, err := w.api.UploadPartWithContext(
		aws.Context(ctx),
		&s3.UploadPartInput{
			Body:       input.Body,
			Bucket:     input.Bucket,
			Key:        input.Key,
			PartNumber: input.PartNumber,
			UploadId:   input.UploadID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.UploadPartOutput{
		ETag: out.ETag,
	}, nil

}
