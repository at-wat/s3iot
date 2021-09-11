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

package awss3v1

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/awss3v1/internal/moq/s3iface"
)

func TestNew(t *testing.T) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials("id", "secret", ""),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run("NewUploader", func(t *testing.T) {
		u := NewUploader(sess)
		if _, ok := u.API.(*wrapper).api.(*s3.S3); !ok {
			t.Errorf("Base API is expected to be *s3.S3, actually %T", u.API.(*wrapper).api)
		}
	})
	t.Run("NewDownloader", func(t *testing.T) {
		d := NewDownloader(sess)
		if _, ok := d.API.(*wrapper).api.(*s3.S3); !ok {
			t.Errorf("Base API is expected to be *s3.S3, actually %T", d.API.(*wrapper).api)
		}
	})
}

func TestWrapper(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
		t.Run("GetObject", func(t *testing.T) {
			r := io.NopCloser(bytes.NewReader([]byte{}))

			api := &s3iface.MockS3API{
				GetObjectWithContextFunc: func(ctx context.Context, input *s3.GetObjectInput, options ...request.Option) (*s3.GetObjectOutput, error) {
					expectStringPtr(t, "Bucket", input.Bucket)
					expectStringPtr(t, "Key", input.Key)
					expectStringPtr(t, "Range", input.Range)
					expectStringPtr(t, "VersionID", input.VersionId)
					return &s3.GetObjectOutput{
						Body:          r,
						ContentType:   aws.String("ContentType"),
						ContentLength: aws.Int64(100),
						ContentRange:  aws.String("ContentRange"),
						ETag:          aws.String("ETag"),
						LastModified:  aws.Time(time.Unix(1, 2)),
						VersionId:     aws.String("VersionID"),
					}, nil
				},
			}
			w := NewAPI(api)
			out, err := w.GetObject(context.TODO(),
				&s3iot.GetObjectInput{
					Bucket:    aws.String("Bucket"),
					Key:       aws.String("Key"),
					Range:     aws.String("Range"),
					VersionID: aws.String("VersionID"),
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if n := len(api.GetObjectWithContextCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			if out.Body != r {
				t.Error("Body reader differs")
			}
			expectStringPtr(t, "ContentType", out.ContentType)
			expectInt64Ptr(t, 100, out.ContentLength)
			expectStringPtr(t, "ContentRange", out.ContentRange)
			expectStringPtr(t, "ETag", out.ETag)
			if !out.LastModified.Equal(time.Unix(1, 2)) {
				t.Error("LastModified differs")
			}
			expectStringPtr(t, "VersionID", out.VersionID)
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
	})
	t.Run("Error", func(t *testing.T) {
		errDummy := errors.New("error")

		t.Run("PutObject", func(t *testing.T) {
			api := &s3iface.MockS3API{
				PutObjectWithContextFunc: func(ctx context.Context, input *s3.PutObjectInput, options ...request.Option) (*s3.PutObjectOutput, error) {
					return nil, errDummy
				},
			}
			w := NewAPI(api)
			if _, err := w.PutObject(context.TODO(), &s3iot.PutObjectInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.PutObjectWithContextCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("CreateMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				CreateMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.CreateMultipartUploadInput, options ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := NewAPI(api)
			if _, err := w.CreateMultipartUpload(context.TODO(), &s3iot.CreateMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.CreateMultipartUploadWithContextCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("CompleteMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				CompleteMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.CompleteMultipartUploadInput, options ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := NewAPI(api)
			if _, err := w.CompleteMultipartUpload(context.TODO(), &s3iot.CompleteMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.CompleteMultipartUploadWithContextCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("AbortMultipartUpload", func(t *testing.T) {
			api := &s3iface.MockS3API{
				AbortMultipartUploadWithContextFunc: func(ctx context.Context, input *s3.AbortMultipartUploadInput, options ...request.Option) (*s3.AbortMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := NewAPI(api)
			if _, err := w.AbortMultipartUpload(context.TODO(), &s3iot.AbortMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.AbortMultipartUploadWithContextCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("UploadPart", func(t *testing.T) {
			api := &s3iface.MockS3API{
				UploadPartWithContextFunc: func(ctx context.Context, input *s3.UploadPartInput, options ...request.Option) (*s3.UploadPartOutput, error) {
					return nil, errDummy
				},
			}
			w := NewAPI(api)
			if _, err := w.UploadPart(context.TODO(), &s3iot.UploadPartInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.UploadPartWithContextCalls()); n != 1 {
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

func expectInt64Ptr(t *testing.T, expected int64, ptr *int64) {
	t.Helper()
	if expected != *ptr {
		t.Errorf("Expected %d, got %d", expected, *ptr)
	}
}
