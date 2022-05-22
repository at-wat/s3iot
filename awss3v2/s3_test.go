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
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/at-wat/s3iot/awss3v2"
	mock_awss3v2 "github.com/at-wat/s3iot/awss3v2/internal/moq/awss3v2"
	mock_s3iface "github.com/at-wat/s3iot/awss3v2/internal/moq/s3iface"
	"github.com/at-wat/s3iot/s3api"
)

func TestWrapper(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("PutObject", func(t *testing.T) {
			r := bytes.NewReader([]byte{})

			api := &mock_awss3v2.MockS3API{
				PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectString(t, "ACL", string(params.ACL))
					expectStringPtr(t, "ContentType", params.ContentType)
					if params.Body != r {
						t.Error("Body reader differs")
					}

					opts := &s3.Options{
						HTTPClient: &mock_s3iface.MockHTTPClient{
							DoFunc: func(request *http.Request) (*http.Response, error) {
								return &http.Response{
									Request: request,
								}, nil
							},
						},
					}
					for _, o := range optFns {
						o(opts)
					}
					opts.HTTPClient.Do(&http.Request{
						URL: &url.URL{Scheme: "s3", Host: "url"},
					})

					return &s3.PutObjectOutput{
						VersionId: aws.String("VersionID"),
						ETag:      aws.String("ETag"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
			out, err := w.PutObject(context.TODO(),
				&s3api.PutObjectInput{
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
			expectStringPtr(t, "s3://url", out.Location)
		})
		t.Run("GetObject", func(t *testing.T) {
			r := io.NopCloser(bytes.NewReader([]byte{}))

			api := &mock_awss3v2.MockS3API{
				GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectStringPtr(t, "Range", params.Range)
					expectStringPtr(t, "VersionID", params.VersionId)
					return &s3.GetObjectOutput{
						Body:          r,
						ContentType:   aws.String("ContentType"),
						ContentLength: 100,
						ContentRange:  aws.String("ContentRange"),
						ETag:          aws.String("ETag"),
						LastModified:  aws.Time(time.Unix(1, 2)),
						VersionId:     aws.String("VersionID"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
			out, err := w.GetObject(context.TODO(),
				&s3api.GetObjectInput{
					Bucket:    aws.String("Bucket"),
					Key:       aws.String("Key"),
					Range:     aws.String("Range"),
					VersionID: aws.String("VersionID"),
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if n := len(api.GetObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			if out.Body != r {
				t.Error("Body reader differs")
			}
			expectStringPtr(t, "ContentType", out.ContentType)
			if *out.ContentLength != 100 {
				t.Error("ContentLength differs")
			}
			expectStringPtr(t, "ContentRange", out.ContentRange)
			expectStringPtr(t, "ETag", out.ETag)
			if !out.LastModified.Equal(time.Unix(1, 2)) {
				t.Error("LastModified differs")
			}
			expectStringPtr(t, "VersionID", out.VersionID)
		})

		t.Run("CreateMultipartUpload", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
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
				&s3api.CreateMultipartUploadInput{
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
			api := &mock_awss3v2.MockS3API{
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
						Location:  aws.String("s3://url"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
			out, err := w.CompleteMultipartUpload(context.TODO(),
				&s3api.CompleteMultipartUploadInput{
					Bucket: aws.String("Bucket"),
					Key:    aws.String("Key"),
					CompletedParts: []*s3api.CompletedPart{
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
			expectStringPtr(t, "s3://url", out.Location)
		})
		t.Run("AbortMultipartUpload", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				AbortMultipartUploadFunc: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
					expectStringPtr(t, "Bucket", params.Bucket)
					expectStringPtr(t, "Key", params.Key)
					expectStringPtr(t, "UploadID", params.UploadId)
					return &s3.AbortMultipartUploadOutput{}, nil
				},
			}
			w := awss3v2.NewAPI(api)
			_, err := w.AbortMultipartUpload(context.TODO(),
				&s3api.AbortMultipartUploadInput{
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

			api := &mock_awss3v2.MockS3API{
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
				&s3api.UploadPartInput{
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
		t.Run("DeleteObject", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				DeleteObjectFunc: func(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
					expectStringPtr(t, "Bucket", input.Bucket)
					expectStringPtr(t, "Key", input.Key)
					expectStringPtr(t, "VersionID", input.VersionId)
					return &s3.DeleteObjectOutput{
						VersionId: aws.String("VersionID2"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
			out, err := w.DeleteObject(context.TODO(),
				&s3api.DeleteObjectInput{
					Bucket:    aws.String("Bucket"),
					Key:       aws.String("Key"),
					VersionID: aws.String("VersionID"),
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if n := len(api.DeleteObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			expectStringPtr(t, "VersionID2", out.VersionID)
		})
		t.Run("ListObjects", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				ListObjectsV2Func: func(ctx context.Context, input *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					expectStringPtr(t, "Bucket", input.Bucket)
					expectStringPtr(t, "ContinuationToken", input.ContinuationToken)
					if input.MaxKeys != 2 {
						t.Errorf("Expected MaxKeys: 2, got: %d", input.MaxKeys)
					}
					expectStringPtr(t, "Prefix", input.Prefix)
					return &s3.ListObjectsV2Output{
						Contents: []types.Object{
							{
								ETag:         aws.String("Etag1"),
								Key:          aws.String("Key1"),
								LastModified: aws.Time(time.Unix(100, 0)),
								Size:         1000,
							},
							{
								ETag: aws.String("Etag2"),
								Key:  aws.String("Key2"),
							},
						},
						KeyCount:              2,
						NextContinuationToken: aws.String("NextToken"),
					}, nil
				},
			}
			w := awss3v2.NewAPI(api)
			out, err := w.ListObjectsV2(context.TODO(),
				&s3api.ListObjectsV2Input{
					Bucket:            aws.String("Bucket"),
					ContinuationToken: aws.String("ContinuationToken"),
					MaxKeys:           2,
					Prefix:            aws.String("Prefix"),
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if n := len(api.ListObjectsV2Calls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
			expectedContents := []s3api.Object{
				{
					ETag:         aws.String("Etag1"),
					Key:          aws.String("Key1"),
					LastModified: aws.Time(time.Unix(100, 0)),
					Size:         1000,
				},
				{
					ETag: aws.String("Etag2"),
					Key:  aws.String("Key2"),
				},
			}
			if !reflect.DeepEqual(expectedContents, out.Contents) {
				t.Errorf("Expected Contents: %v, got: %v", expectedContents, out.Contents)
			}
			if out.KeyCount != 2 {
				t.Errorf("Expected KeyCount: 2, got: %d", out.KeyCount)
			}
			expectStringPtr(t, "NextToken", out.NextContinuationToken)
		})
	})
	t.Run("Error", func(t *testing.T) {
		errDummy := errors.New("error")

		t.Run("PutObject", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.PutObject(context.TODO(), &s3api.PutObjectInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.PutObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("GetObject", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.GetObject(context.TODO(), &s3api.GetObjectInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.GetObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("CreateMultipartUpload", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				CreateMultipartUploadFunc: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.CreateMultipartUpload(context.TODO(), &s3api.CreateMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.CreateMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("CompleteMultipartUpload", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				CompleteMultipartUploadFunc: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.CompleteMultipartUpload(context.TODO(), &s3api.CompleteMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.CompleteMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("AbortMultipartUpload", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				AbortMultipartUploadFunc: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.AbortMultipartUpload(context.TODO(), &s3api.AbortMultipartUploadInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.AbortMultipartUploadCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("UploadPart", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				UploadPartFunc: func(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.UploadPart(context.TODO(), &s3api.UploadPartInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.UploadPartCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("DeleteObject", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				DeleteObjectFunc: func(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.DeleteObject(context.TODO(), &s3api.DeleteObjectInput{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.DeleteObjectCalls()); n != 1 {
				t.Errorf("Expected calls: 1, actual: %d", n)
			}
		})
		t.Run("ListObjects", func(t *testing.T) {
			api := &mock_awss3v2.MockS3API{
				ListObjectsV2Func: func(ctx context.Context, input *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					return nil, errDummy
				},
			}
			w := awss3v2.NewAPI(api)
			if _, err := w.ListObjectsV2(context.TODO(), &s3api.ListObjectsV2Input{}); err != errDummy {
				t.Fatal("Expected error")
			}
			if n := len(api.ListObjectsV2Calls()); n != 1 {
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
