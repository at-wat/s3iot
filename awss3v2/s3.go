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
	"github.com/at-wat/s3iot/awss3v2/internal/locationstore"
	"github.com/at-wat/s3iot/s3api"

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

// NewAPI wraps s3.Client to s3api.S3API.
func NewAPI(api S3API) s3api.S3API {
	return &wrapper{api: api}
}

type wrapper struct {
	api S3API
}

func (w *wrapper) PutObject(ctx context.Context, input *s3api.PutObjectInput) (*s3api.PutObjectOutput, error) {
	var acl s3types.ObjectCannedACL
	if input.ACL != nil {
		acl = s3types.ObjectCannedACL(*input.ACL)
	}
	ls := &locationstore.LocationStore{}
	out, err := w.api.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      input.Bucket,
			Key:         input.Key,
			ACL:         acl,
			Body:        input.Body,
			ContentType: input.ContentType,
		},
		func(o *s3.Options) {
			ls.HTTPClient, o.HTTPClient = o.HTTPClient, ls
		},
	)
	if err != nil {
		return nil, err
	}
	return &s3api.PutObjectOutput{
		VersionID: out.VersionId,
		ETag:      out.ETag,
		Location:  &ls.Location,
	}, nil
}

func (w *wrapper) GetObject(ctx context.Context, input *s3api.GetObjectInput) (*s3api.GetObjectOutput, error) {
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
	return &s3api.GetObjectOutput{
		Body:          out.Body,
		ContentType:   out.ContentType,
		ContentLength: &out.ContentLength,
		ContentRange:  out.ContentRange,
		ETag:          out.ETag,
		LastModified:  out.LastModified,
		VersionID:     out.VersionId,
	}, nil
}

func (w *wrapper) CreateMultipartUpload(ctx context.Context, input *s3api.CreateMultipartUploadInput) (*s3api.CreateMultipartUploadOutput, error) {
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
	return &s3api.CreateMultipartUploadOutput{
		UploadID: out.UploadId,
	}, nil
}

func (w *wrapper) CompleteMultipartUpload(ctx context.Context, input *s3api.CompleteMultipartUploadInput) (*s3api.CompleteMultipartUploadOutput, error) {
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
	return &s3api.CompleteMultipartUploadOutput{
		VersionID: out.VersionId,
		ETag:      out.ETag,
		Location:  out.Location,
	}, nil
}

func (w *wrapper) AbortMultipartUpload(ctx context.Context, input *s3api.AbortMultipartUploadInput) (*s3api.AbortMultipartUploadOutput, error) {
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
	return &s3api.AbortMultipartUploadOutput{}, nil
}

func (w *wrapper) UploadPart(ctx context.Context, input *s3api.UploadPartInput) (*s3api.UploadPartOutput, error) {
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
	return &s3api.UploadPartOutput{
		ETag: out.ETag,
	}, nil
}

func (w *wrapper) DeleteObject(ctx context.Context, input *s3api.DeleteObjectInput) (*s3api.DeleteObjectOutput, error) {
	out, err := w.api.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket:    input.Bucket,
			Key:       input.Key,
			VersionId: input.VersionID,
		})
	if err != nil {
		return nil, err
	}
	return &s3api.DeleteObjectOutput{
		VersionID: out.VersionId,
	}, nil
}

func (w *wrapper) ListObjects(ctx context.Context, input *s3api.ListObjectsInput) (*s3api.ListObjectsOutput, error) {
	out, err := w.api.ListObjectsV2(
		ctx,
		&s3.ListObjectsV2Input{
			Bucket:            input.Bucket,
			ContinuationToken: input.ContinuationToken,
			MaxKeys:           int32(input.MaxKeys),
			Prefix:            input.Prefix,
		})
	if err != nil {
		return nil, err
	}
	contents := make([]s3api.Object, len(out.Contents))
	for i, c := range out.Contents {
		contents[i] = s3api.Object{
			ETag:         c.ETag,
			Key:          c.Key,
			LastModified: c.LastModified,
			Size:         c.Size,
		}
	}
	return &s3api.ListObjectsOutput{
		Contents:              contents,
		KeyCount:              int(out.KeyCount),
		NextContinuationToken: out.NextContinuationToken,
	}, nil
}
