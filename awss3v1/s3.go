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
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/s3api"
)

// NewUploader creates s3iot.Uploader from aws-sdk-go ConfigProvider (like Session).
func NewUploader(c client.ConfigProvider, opts ...s3iot.UploaderOption) *s3iot.Uploader {
	u := &s3iot.Uploader{
		UpDownloaderBase: s3iot.UpDownloaderBase{
			API:             NewAPI(s3.New(c)),
			ErrorClassifier: &ErrorClassifier{},
		},
	}
	for _, opt := range opts {
		opt.ApplyToUploader(u)
	}
	return u
}

// NewDownloader creates s3iot.Downloader from aws-sdk-go ConfigProvider (like Session).
func NewDownloader(c client.ConfigProvider, opts ...s3iot.DownloaderOption) *s3iot.Downloader {
	d := &s3iot.Downloader{
		UpDownloaderBase: s3iot.UpDownloaderBase{
			API: NewAPI(s3.New(c)),
		},
	}
	for _, opt := range opts {
		opt.ApplyToDownloader(d)
	}
	return d
}

// NewAPI wraps s3iface.S3API to s3api.UpDownloadAPI.
func NewAPI(api s3iface.S3API) s3api.UpDownloadAPI {
	return &wrapper{api: api}
}

type wrapper struct {
	api s3iface.S3API
}

func (w *wrapper) PutObject(ctx context.Context, input *s3api.PutObjectInput) (*s3api.PutObjectOutput, error) {
	var req *request.Request
	out, err := w.api.PutObjectWithContext(
		aws.Context(ctx),
		&s3.PutObjectInput{
			Bucket:      input.Bucket,
			Key:         input.Key,
			ACL:         input.ACL,
			Body:        input.Body,
			ContentType: input.ContentType,
		}, func(r *request.Request) {
			req = r
		},
	)
	if err != nil {
		return nil, err
	}
	u := *req.HTTPResponse.Request.URL
	u.RawQuery = ""
	location := u.String()
	return &s3api.PutObjectOutput{
		VersionID: out.VersionId,
		ETag:      out.ETag,
		Location:  &location,
	}, nil
}

func (w *wrapper) GetObject(ctx context.Context, input *s3api.GetObjectInput) (*s3api.GetObjectOutput, error) {
	out, err := w.api.GetObjectWithContext(
		aws.Context(ctx),
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
		ContentLength: out.ContentLength,
		ContentRange:  out.ContentRange,
		ETag:          out.ETag,
		LastModified:  out.LastModified,
		VersionID:     out.VersionId,
	}, nil
}

func (w *wrapper) CreateMultipartUpload(ctx context.Context, input *s3api.CreateMultipartUploadInput) (*s3api.CreateMultipartUploadOutput, error) {
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
	return &s3api.CreateMultipartUploadOutput{
		UploadID: out.UploadId,
	}, nil
}

func (w *wrapper) CompleteMultipartUpload(ctx context.Context, input *s3api.CompleteMultipartUploadInput) (*s3api.CompleteMultipartUploadOutput, error) {
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
	return &s3api.CompleteMultipartUploadOutput{
		VersionID: out.VersionId,
		ETag:      out.ETag,
		Location:  out.Location,
	}, nil
}

func (w *wrapper) AbortMultipartUpload(ctx context.Context, input *s3api.AbortMultipartUploadInput) (*s3api.AbortMultipartUploadOutput, error) {
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
	return &s3api.AbortMultipartUploadOutput{}, nil
}

func (w *wrapper) UploadPart(ctx context.Context, input *s3api.UploadPartInput) (*s3api.UploadPartOutput, error) {
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
	return &s3api.UploadPartOutput{
		ETag: out.ETag,
	}, nil

}
