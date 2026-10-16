package awss3v2

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"

	"github.com/at-wat/s3iot"
	mock_transfermanageriface "github.com/at-wat/s3iot/awss3v2/internal/moq/transfermanageriface"
)

func TestNewAWSSDKUploader(t *testing.T) {
	ctx := context.Background()

	testCases := map[string]struct {
		expectedInput  tm.UploadObjectInput
		expectedOutput s3iot.UploadOutput
		expectedErr    error
	}{
		"Normal": {
			expectedInput: tm.UploadObjectInput{
				ACL:         types.ObjectCannedACLPrivate,
				Bucket:      aws.String("bucket"),
				ContentType: aws.String("content-type"),
				Key:         aws.String("key"),
			},
			expectedOutput: s3iot.UploadOutput{
				ETag:      aws.String("etag"),
				VersionID: aws.String("versionID"),
				Location:  aws.String("location"),
			},
		},
		"NilACL": {
			expectedInput: tm.UploadObjectInput{
				Bucket:      aws.String("bucket"),
				ContentType: aws.String("content-type"),
				Key:         aws.String("key"),
			},
			expectedOutput: s3iot.UploadOutput{
				ETag:      aws.String("etag"),
				VersionID: aws.String("versionID"),
				Location:  aws.String("location"),
			},
		},
		"Error": {
			expectedInput: tm.UploadObjectInput{
				Bucket:      aws.String("bucket"),
				ContentType: aws.String("content-type"),
				Key:         aws.String("key"),
			},
			expectedErr: errors.New("a error"),
		},
	}
	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {

			api := &mock_transfermanageriface.MockClient{
				UploadObjectFunc: func(uCtx context.Context, input *tm.UploadObjectInput, opts ...func(*tm.Options)) (*tm.UploadObjectOutput, error) {
					if !reflect.DeepEqual(tt.expectedInput, *input) {
						t.Errorf("Expected: %v, got: %v", tt.expectedInput, *input)
					}
					if tt.expectedErr != nil {
						return nil, tt.expectedErr
					}
					return &tm.UploadObjectOutput{
						ETag:      tt.expectedOutput.ETag,
						VersionID: tt.expectedOutput.VersionID,
						Location:  tt.expectedOutput.Location,
					}, nil
				},
			}
			u := NewAWSSDKUploader(api)
			uc, err := u.Upload(ctx, &s3iot.UploadInput{
				ACL:         aws.String(string(tt.expectedInput.ACL)),
				Bucket:      tt.expectedInput.Bucket,
				ContentType: tt.expectedInput.ContentType,
				Key:         tt.expectedInput.Key,
			})
			if err != nil {
				t.Fatal(err)
			}
			uc.Pause()
			uc.Resume()
			<-uc.Done()

			if _, err := uc.Status(); tt.expectedErr != err {
				t.Errorf("Expected error '%v', got '%v'", tt.expectedErr, err)
			}
			res, err := uc.Result()
			if tt.expectedErr != err {
				t.Errorf("Expected error '%v', got '%v'", tt.expectedErr, err)
			}
			if !reflect.DeepEqual(tt.expectedOutput, res) {
				t.Errorf("Expected: %v, got: %v", tt.expectedOutput, res)
			}
			if n := len(api.UploadObjectCalls()); n != 1 {
				t.Fatalf("Expected 1 call, called %d", n)
			}
		})
	}
}
