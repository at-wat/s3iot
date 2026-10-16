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
	mock_s3manageriface "github.com/at-wat/s3iot/awss3v2-s3manager/internal/moq/s3manageriface"
	"github.com/at-wat/s3iot/internal/iotest"
)

func TestNewAWSSDKDownloader(t *testing.T) {
	const expectedLen = 100

	for name, testCase := range map[string]struct {
		expectedStatus s3iot.DownloadStatus
		expectedErr    error
	}{
		"Normal": {
			expectedStatus: s3iot.DownloadStatus{
				Status: s3iot.Status{
					CompletedSize: expectedLen,
					Size:          expectedLen,
				},
			},
			expectedErr: nil,
		},
		"Error": {
			expectedStatus: s3iot.DownloadStatus{
				Status: s3iot.Status{
					CompletedSize: 10,
					Size:          0,
				},
			},
			expectedErr: errors.New("a error"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			expectedInput := s3.GetObjectInput{
				Bucket:    aws.String("bucket"),
				Key:       aws.String("key"),
				VersionId: aws.String("version-id"),
			}
			expectedOutput := s3iot.DownloadOutput{
				VersionID: aws.String("version-id"),
			}
			expectedWriter := &iotest.BufferAt{}

			api := &mock_s3manageriface.MockDownloader{
				DownloadFunc: func(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, opts ...func(*manager.Downloader)) (int64, error) {
					if w != expectedWriter {
						t.Errorf("Unexpected writer")
					}
					if !reflect.DeepEqual(expectedInput, *input) {
						t.Errorf("Expected: %v, got: %v", expectedInput, *input)
					}
					if testCase.expectedErr != nil {
						return 10, testCase.expectedErr
					}
					return expectedLen, nil
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
			if testCase.expectedErr != err {
				t.Errorf("Expected error '%v', got '%v'", testCase.expectedErr, err)
			}
			if !reflect.DeepEqual(testCase.expectedStatus, st) {
				t.Errorf("Expected: %v, got: %v", testCase.expectedStatus, st)
			}
			res, err := uc.Result()
			if testCase.expectedErr != err {
				t.Errorf("Expected error '%v', got '%v'", testCase.expectedErr, err)
			}
			if !reflect.DeepEqual(expectedOutput, res) {
				t.Errorf("Expected: %v, got: %v", expectedOutput, res)
			}
			if n := len(api.DownloadCalls()); n != 1 {
				t.Fatalf("Expected 1 call, called %d", n)
			}
		})
	}
}
