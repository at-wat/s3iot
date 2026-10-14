package awss3v2

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/at-wat/s3iot"
	mock_s3manageriface "github.com/at-wat/s3iot/awss3v2-1.22.2/internal/moq/s3manageriface"
	"github.com/at-wat/s3iot/internal/iotest"
)

func TestNewAWSSDKDownloader(t *testing.T) {
	ctx := context.Background()
	expectedInput := s3.GetObjectInput{
		Bucket:    aws.String("bucket"),
		Key:       aws.String("key"),
		VersionId: aws.String("version-id"),
	}
	expectedLen := int64(100)
	expectedOutput := s3iot.DownloadOutput{
		VersionID: aws.String("version-id"),
	}
	expectedStatus := s3iot.DownloadStatus{
		Status: s3iot.Status{
			CompletedSize: expectedLen,
		},
	}
	expectedErr := errors.New("a error")
	expectedWriter := iotest.BufferAt{}

	api := &mock_s3manageriface.MockDownloader{
		DownloadFunc: func(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, opts ...func(*manager.Downloader)) (int64, error) {
			if !reflect.DeepEqual(expectedInput, *input) {
				t.Errorf("Expected: %v, got: %v", expectedInput, *input)
			}
			return expectedLen, expectedErr
		},
	}
	u := NewAWSSDKDownloader(api)
	uc, err := u.Download(ctx, expectedWriter, &s3iot.DownloadInput{
		Bucket:    expectedInput.Bucket,
		Key:       expectedInput.Key,
		VersionID: expectedInput.VersionId,
	})
	if err != nil {
		t.Fatal(err)
	}
	uc.Pause()
	uc.Resume()
	<-uc.Done()

	st, err := uc.Status()
	if expectedErr != err {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
	if !reflect.DeepEqual(expectedStatus, st) {
		t.Errorf("Expected: %v, got: %v", expectedStatus, st)
	}
	res, err := uc.Result()
	if expectedErr != err {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
	if !reflect.DeepEqual(expectedOutput, res) {
		t.Errorf("Expected: %v, got: %v", expectedOutput, res)
	}
	if n := len(api.DownloadCalls()); n != 1 {
		t.Fatalf("Expected 1 call, called %d", n)
	}
}
