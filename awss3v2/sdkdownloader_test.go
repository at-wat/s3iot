package awss3v2

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"

	"github.com/at-wat/s3iot"
	mock_transfermanageriface "github.com/at-wat/s3iot/awss3v2/internal/moq/transfermanageriface"
	"github.com/at-wat/s3iot/internal/iotest"
)

func TestNewAWSSDKDownloader(t *testing.T) {
	ctx := context.Background()
	expectedInput := tm.DownloadObjectInput{
		Bucket:    aws.String("bucket"),
		Key:       aws.String("key"),
		VersionID: aws.String("version-id"),
	}
	expectedLen := int64(100)
	expectedOutput := s3iot.DownloadOutput{
		VersionID: aws.String("version-id"),
	}
	expectedStatus := s3iot.DownloadStatus{
		Status: s3iot.Status{
			CompletedSize: expectedLen,
			Size:          expectedLen,
		},
	}
	expectedErr := errors.New("a error")
	expectedWriter := &iotest.BufferAt{}

	api := &mock_transfermanageriface.MockClient{
		DownloadObjectFunc: func(ctx context.Context, input *tm.DownloadObjectInput, opts ...func(*tm.Options)) (*tm.DownloadObjectOutput, error) {
			if input.WriterAt != expectedWriter {
				t.Errorf("Unexpected writer")
			}
			input.WriterAt = nil
			if !reflect.DeepEqual(expectedInput, *input) {
				t.Errorf("Expected: %v, got: %v", expectedInput, *input)
			}
			return &tm.DownloadObjectOutput{
				ContentLength: aws.Int64(expectedLen),
			}, expectedErr
		},
	}
	u := NewAWSSDKDownloader(api)
	uc, err := u.Download(ctx, expectedWriter, &s3iot.DownloadInput{
		Bucket:    expectedInput.Bucket,
		Key:       expectedInput.Key,
		VersionID: expectedInput.VersionID,
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
	if n := len(api.DownloadObjectCalls()); n != 1 {
		t.Fatalf("Expected 1 call, called %d", n)
	}
}
