package awss3v1

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/at-wat/s3iot"
	mock_s3manageriface "github.com/at-wat/s3iot/awss3v1/internal/moq/s3manageriface"
)

func TestNewAWSSDKUploader(t *testing.T) {
	ctx := context.Background()
	expectedInput := s3manager.UploadInput{
		ACL:         aws.String("acl"),
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
		UploadWithContextFunc: func(uCtx context.Context, input *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			if !reflect.DeepEqual(expectedInput, *input) {
				t.Errorf("Expected: %v, got: %v", expectedInput, *input)
			}
			return &s3manager.UploadOutput{
				ETag:      expectedOutput.ETag,
				VersionID: expectedOutput.VersionID,
				Location:  *expectedOutput.Location,
			}, expectedErr
		},
	}
	u := NewAWSSDKUploader(api)
	uc, err := u.Upload(ctx, &s3iot.UploadInput{
		ACL:         expectedInput.ACL,
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
	if n := len(api.UploadWithContextCalls()); n != 1 {
		t.Fatalf("Expected 1 call, called %d", n)
	}
}
