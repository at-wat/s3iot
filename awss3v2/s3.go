package awss3v2

import (
	"context"

	"github.com/at-wat/s3iot"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func NewUploader(c aws.Config, opts ...s3iot.UploaderOption) *s3iot.Uploader {
	u := &s3iot.Uploader{
		API: NewAPI(s3.NewFromConfig(c)),
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

func NewAPI(api *s3.Client) s3iot.S3API {
	return &wrapper{api: api}
}

type wrapper struct {
	api *s3.Client
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
	out, err := w.api.UploadPart(
		ctx,
		&s3.UploadPartInput{
			Body:       input.Body,
			Bucket:     input.Bucket,
			Key:        input.Key,
			PartNumber: int32(*input.PartNumber),
			UploadId:   input.UploadID,
		})
	if err != nil {
		return nil, err
	}
	return &s3iot.UploadPartOutput{
		ETag: out.ETag,
	}, nil

}
