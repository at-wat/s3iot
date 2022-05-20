package awssdkv2

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/at-wat/s3iot"
	mock_s3manageriface "github.com/at-wat/s3iot/awssdkv2/internal/moq/s3manageriface"
)

func TestNewAWSSDKUploader(t *testing.T) {
	ctx := context.Background()
	expectedInput := s3.PutObjectInput{
		ACL:         types.ObjectCannedACLPrivate,
		Bucket:      aws.String("bucket"),
		ContentType: aws.String("content-type"),
		Key:         aws.String("key"),
	}
	expectedOutput := s3iot.UploadOutput{
		ETag:      aws.String("etag"),
		VersionID: aws.String("versionID"),
		Location:  aws.String("location"),
	}
	expectedErr := errors.New("a error")

	api := &mock_s3manageriface.MockUploader{
		UploadFunc: func(uCtx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
			if !reflect.DeepEqual(expectedInput, *input) {
				t.Errorf("Expected: %v, got: %v", expectedInput, *input)
			}
			return &manager.UploadOutput{
				ETag:      expectedOutput.ETag,
				VersionID: expectedOutput.VersionID,
				Location:  *expectedOutput.Location,
			}, expectedErr
		},
	}
	u := NewAWSSDKUploader(api)
	uc, err := u.Upload(ctx, &s3iot.UploadInput{
		ACL:         aws.String(string(expectedInput.ACL)),
		Bucket:      expectedInput.Bucket,
		ContentType: expectedInput.ContentType,
		Key:         expectedInput.Key,
	})
	if err != nil {
		t.Fatal(err)
	}
	uc.Pause()
	uc.Resume()
	<-uc.Done()

	if _, err := uc.Status(); expectedErr != err {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
	res, err := uc.Result()
	if expectedErr != err {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
	if !reflect.DeepEqual(expectedOutput, res) {
		t.Errorf("Expected: %v, got: %v", expectedOutput, res)
	}
	if n := len(api.UploadCalls()); n != 1 {
		t.Fatalf("Expected 1 call, called %d", n)
	}
}
