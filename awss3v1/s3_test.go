package awss3v1

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/awss3v1/internal/moq/s3iface"
)

func TestWrapper(t *testing.T) {
	t.Run("PutObject", func(t *testing.T) {
		r := bytes.NewReader([]byte{})

		api := &s3iface.MockS3API{
			PutObjectWithContextFunc: func(ctx context.Context, input *s3.PutObjectInput, options ...request.Option) (*s3.PutObjectOutput, error) {
				expectStringPtr(t, "Bucket", input.Bucket)
				expectStringPtr(t, "Key", input.Key)
				expectStringPtr(t, "ACL", input.ACL)
				expectStringPtr(t, "ContentType", input.ContentType)
				if input.Body != r {
					t.Error("Body reader differs")
				}
				return &s3.PutObjectOutput{
					VersionId: aws.String("VersionID"),
					ETag:      aws.String("ETag"),
				}, nil
			},
		}
		w := NewAPI(api)
		out, err := w.PutObject(context.TODO(),
			&s3iot.PutObjectInput{
				Bucket:      aws.String("Bucket"),
				Key:         aws.String("Key"),
				ACL:         aws.String("ACL"),
				Body:        r,
				ContentType: aws.String("ContentType"),
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(api.PutObjectWithContextCalls()); n != 1 {
			t.Errorf("Expected calls: 1, actual: %d", n)
		}
		expectStringPtr(t, "VersionID", out.VersionID)
		expectStringPtr(t, "ETag", out.ETag)
	})
	t.Run("CreateMultipartUpload", func(t *testing.T) {
		api := &s3iface.MockS3API{
			CreateMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.CreateMultipartUploadInput, options ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
				expectStringPtr(t, "Bucket", input.Bucket)
				expectStringPtr(t, "Key", input.Key)
				expectStringPtr(t, "ACL", input.ACL)
				expectStringPtr(t, "ContentType", input.ContentType)
				return &s3.CreateMultipartUploadOutput{
					UploadId: aws.String("UploadID"),
				}, nil
			},
		}
		w := NewAPI(api)
		out, err := w.CreateMultipartUpload(context.TODO(),
			&s3iot.CreateMultipartUploadInput{
				Bucket:      aws.String("Bucket"),
				Key:         aws.String("Key"),
				ACL:         aws.String("ACL"),
				ContentType: aws.String("ContentType"),
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(api.CreateMultipartUploadWithContextCalls()); n != 1 {
			t.Errorf("Expected calls: 1, actual: %d", n)
		}
		expectStringPtr(t, "UploadID", out.UploadID)
	})
	t.Run("CompleteMultipartUpload", func(t *testing.T) {
		api := &s3iface.MockS3API{
			CompleteMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.CompleteMultipartUploadInput, options ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
				expectStringPtr(t, "Bucket", input.Bucket)
				expectStringPtr(t, "Key", input.Key)
				expectStringPtr(t, "UploadID", input.UploadId)
				expectStringPtr(t, "ETag1", input.MultipartUpload.Parts[0].ETag)
				expectInt64Ptr(t, 1, input.MultipartUpload.Parts[0].PartNumber)
				expectStringPtr(t, "ETag2", input.MultipartUpload.Parts[1].ETag)
				expectInt64Ptr(t, 2, input.MultipartUpload.Parts[1].PartNumber)
				return &s3.CompleteMultipartUploadOutput{
					VersionId: aws.String("VersionID"),
					ETag:      aws.String("ETag"),
				}, nil
			},
		}
		w := NewAPI(api)
		out, err := w.CompleteMultipartUpload(context.TODO(),
			&s3iot.CompleteMultipartUploadInput{
				Bucket: aws.String("Bucket"),
				Key:    aws.String("Key"),
				CompletedParts: []*s3iot.CompletedPart{
					{
						ETag:       aws.String("ETag1"),
						PartNumber: aws.Int64(1),
					},
					{
						ETag:       aws.String("ETag2"),
						PartNumber: aws.Int64(2),
					},
				},
				UploadID: aws.String("UploadID"),
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(api.CompleteMultipartUploadWithContextCalls()); n != 1 {
			t.Errorf("Expected calls: 1, actual: %d", n)
		}
		expectStringPtr(t, "VersionID", out.VersionID)
		expectStringPtr(t, "ETag", out.ETag)
	})
	t.Run("AbortMultipartUpload", func(t *testing.T) {
		api := &s3iface.MockS3API{
			AbortMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.AbortMultipartUploadInput, options ...request.Option) (*s3.AbortMultipartUploadOutput, error) {
				expectStringPtr(t, "Bucket", input.Bucket)
				expectStringPtr(t, "Key", input.Key)
				expectStringPtr(t, "UploadID", input.UploadId)
				return &s3.AbortMultipartUploadOutput{}, nil
			},
		}
		w := NewAPI(api)
		_, err := w.AbortMultipartUpload(context.TODO(),
			&s3iot.AbortMultipartUploadInput{
				Bucket:   aws.String("Bucket"),
				Key:      aws.String("Key"),
				UploadID: aws.String("UploadID"),
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(api.AbortMultipartUploadWithContextCalls()); n != 1 {
			t.Errorf("Expected calls: 1, actual: %d", n)
		}
	})
	t.Run("UploadPart", func(t *testing.T) {
		r := bytes.NewReader([]byte{})

		api := &s3iface.MockS3API{
			UploadPartWithContextFunc: func(ctx context.Context, input *s3.UploadPartInput, options ...request.Option) (*s3.UploadPartOutput, error) {
				if input.Body != r {
					t.Error("Body reader differs")
				}
				expectStringPtr(t, "Bucket", input.Bucket)
				expectStringPtr(t, "Key", input.Key)
				expectInt64Ptr(t, 1, input.PartNumber)
				expectStringPtr(t, "UploadID", input.UploadId)
				return &s3.UploadPartOutput{
					ETag: aws.String("ETag"),
				}, nil
			},
		}
		w := NewAPI(api)
		out, err := w.UploadPart(context.TODO(),
			&s3iot.UploadPartInput{
				Body:       r,
				Bucket:     aws.String("Bucket"),
				Key:        aws.String("Key"),
				PartNumber: aws.Int64(1),
				UploadID:   aws.String("UploadID"),
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(api.UploadPartWithContextCalls()); n != 1 {
			t.Errorf("Expected calls: 1, actual: %d", n)
		}
		expectStringPtr(t, "ETag", out.ETag)
	})

}

func expectStringPtr(t *testing.T, expected string, ptr *string) {
	t.Helper()
	if expected != *ptr {
		t.Errorf("Expected '%s', got '%s'", expected, *ptr)
	}
}

func expectInt64Ptr(t *testing.T, expected int64, ptr *int64) {
	t.Helper()
	if expected != *ptr {
		t.Errorf("Expected %d, got %d", expected, *ptr)
	}
}
