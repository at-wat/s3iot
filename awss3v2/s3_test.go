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

package awss3v2_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/awss3v2"
	"github.com/at-wat/s3iot/awss3v2/internal/moq/s3iface"
)

func TestWrapper(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("PutObject", func(t *testing.T) {
			r := bytes.NewReader([]byte{})

			api := &s3iface.MockS3API{
				PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectString(t, "ACL", string(params.ACL))
					expectStringPtr(t, "ContentType", params.ContentType)
					if params.Body != r {
						t.Error("Body reader differs")
					}
					return &s3.PutObjectOutput{
						VersionId: aws.String("VersionID"),
						ETag:      aws.String("ETag"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
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
			if n := len(api.PutObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			expectStringPtr(t, "VersionID", out.VersionID)
			expectStringPtr(t, "ETag", out.ETag)
		})
		t.Run("CreateMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				CreateMultipartUploadFunc: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectString(t, "ACL", string(params.ACL))
					expectStringPtr(t, "ContentType", params.ContentType)
					return &s3.CreateMultipartUploadOutput{
						UploadId: aws.String("UploadID"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
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
			if n := len(api.CreateMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			expectStringPtr(t, "UploadID", out.UploadID)
		})
		t.Run("CompleteMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				CompleteMultipartUploadFunc: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectStringPtr(t, "UploadID", params.UploadId)
					expectStringPtr(t, "ETag1", params.MultipartUpload.Parts[0].ETag)
					expectInt32(t, 1, params.MultipartUpload.Parts[0].PartNumber)
					expectStringPtr(t, "ETag2", params.MultipartUpload.Parts[1].ETag)
					expectInt32(t, 2, params.MultipartUpload.Parts[1].PartNumber)
					return &s3.CompleteMultipartUploadOutput{
						VersionId: aws.String("VersionID"),
						ETag:      aws.String("ETag"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
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
			if n := len(api.CompleteMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			expectStringPtr(t, "VersionID", out.VersionID)
			expectStringPtr(t, "ETag", out.ETag)
		})
		t.Run("AbortMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				AbortMultipartUploadFunc: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectStringPtr(t, "UploadID", params.UploadId)
					return &s3.AbortMultipartUploadOutput{}, nil
				},
			}
			w := awss3v2.NewAPI(api)
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
			if n := len(api.AbortMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("UploadPart", func(t *testing.T) {
			r := bytes.NewReader([]byte{})

			api := &s3iface.MockS3API{
				UploadPartFunc: func(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
					if params.Body != r {
						t.Error("Body reader differs")
					}
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectInt32(t, 1, params.PartNumber)
					expectStringPtr(t, "UploadID", params.UploadId)
					return &s3.UploadPartOutput{
						ETag: aws.String("ETag"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
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
			if n := len(api.UploadPartCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			expectStringPtr(t, "ETag", out.ETag)
		})
	})
	t.Run("Error", func(t *testing.T) {
		errDummy := errors.New("error")

		t.Run("PutObject", func(t *testing.T) {
			api := &s3iface.MockS3API{
				PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.PutObject(context.TODO(), &s3iot.PutObjectInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.PutObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("CreateMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				CreateMultipartUploadFunc: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.CreateMultipartUpload(context.TODO(), &s3iot.CreateMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.CreateMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("CompleteMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				CompleteMultipartUploadFunc: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.CompleteMultipartUpload(context.TODO(), &s3iot.CompleteMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.CompleteMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("AbortMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				AbortMultipartUploadFunc: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.AbortMultipartUpload(context.TODO(), &s3iot.AbortMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.AbortMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("UploadPart", func(t *testing.T) {
			api := &s3iface.MockS3API{
				UploadPartFunc: func(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.UploadPart(context.TODO(), &s3iot.UploadPartInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.UploadPartCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
	})
}

func expectStringPtr(t *testing.T, expected string, ptr *string) {
	t.Helper()
	if expected != *ptr {
		t.Errorf("Expected '%s', got '%s'", expected, *ptr)
	}
}

func expectString(t *testing.T, expected string, val string) {
	t.Helper()
	if expected != val {
		t.Errorf("Expected '%s', got '%s'", expected, val)
	}
}

func expectInt32(t *testing.T, expected int32, val int32) {
	t.Helper()
	if expected != val {
		t.Errorf("Expected %d, got %d", expected, val)
	}
}
