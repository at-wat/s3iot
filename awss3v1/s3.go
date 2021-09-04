package awss3v1

import (
	"context"

	"github.com/at-wat/s3iot"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func NewUploader(c client.ConfigProvider) *s3iot.Uploader {
	return &s3iot.Uploader{
		API: NewAPI(s3.New(c)),
	}
}

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
