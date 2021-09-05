package s3iot_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/at-wat/s3iot"
	mock_s3iot "github.com/at-wat/s3iot/internal/moq/s3iot"
)

func TestUploader(t *testing.T) {
	var (
		bucket = "Bucket"
		key    = "Key"
	)
	data := make([]byte, 128)
	if _, err := rand.Read(data); err != nil {
		t.Fatal(err)
	}

	t.Run("SinglePart", func(t *testing.T) {
		buf := &bytes.Buffer{}
		api := newMockAPI(buf)
		u := &s3iot.Uploader{}
		s3iot.WithAPI(api)(u)
		s3iot.WithPacketizer(&s3iot.DefaultPacketizerFactory{PartSize: 128})(u)
		s3iot.WithRetryer(&s3iot.NoRetryerFactory{})(u)
		s3iot.WithReadInterceptor(nil)(u)

		uc, err := u.Upload(context.TODO(), &s3iot.UploadInput{
			Bucket: &bucket,
			Key:    &key,
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-uc.Done():
		}
		out, err := uc.Result()
		if err != nil {
			t.Fatal(err)
		}

		if len(api.CreateMultipartUploadCalls()) != 0 {
			t.Fatal("CreateMultipartUpload must not be called")
		}
		if n := len(api.PutObjectCalls()); n != 1 {
			t.Fatalf("PutObject must be called once, but called %d times", n)
		}

		if *out.ETag != "TAG1" {
			t.Errorf("Expected ETag: TAG0, got: %s", *out.ETag)
		}
		if !bytes.Equal(data, buf.Bytes()) {
			t.Error("Uploaded data differs")
		}
	})
	t.Run("MultiPart", func(t *testing.T) {
		buf := &bytes.Buffer{}
		api := newMockAPI(buf)
		u := &s3iot.Uploader{}
		s3iot.WithAPI(api)(u)
		s3iot.WithPacketizer(&s3iot.DefaultPacketizerFactory{PartSize: 50})(u)
		s3iot.WithRetryer(&s3iot.NoRetryerFactory{})(u)
		s3iot.WithReadInterceptor(nil)(u)

		uc, err := u.Upload(context.TODO(), &s3iot.UploadInput{
			Bucket: &bucket,
			Key:    &key,
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-time.After(time.Second):
			t.Fatal("Timeout")
		case <-uc.Done():
		}
		out, err := uc.Result()
		if err != nil {
			t.Fatal(err)
		}

		if n := len(api.CreateMultipartUploadCalls()); n != 1 {
			t.Fatalf("CreateMultipartUpload must be called once, but called %d times", n)
		}
		if n := len(api.UploadPartCalls()); n != 3 {
			t.Fatalf("UploadPart must be called 3 times, but called %d times", n)
		}
		if n := len(api.CompleteMultipartUploadCalls()); n != 1 {
			t.Fatalf("CompleteMultipartUpload must be called once, but called %d times", n)
		}
		if len(api.PutObjectCalls()) != 0 {
			t.Fatal("PutObject must not be called")
		}

		if *out.ETag != "TAG4" {
			t.Errorf("Expected ETag: TAG0, got: %s", *out.ETag)
		}
		if !bytes.Equal(data, buf.Bytes()) {
			t.Error("Uploaded data differs")
		}

		comp := api.CompleteMultipartUploadCalls()
		if n := len(comp[0].Input.CompletedParts); n != 3 {
			t.Fatalf("Expected 3 parts, actually %d parts", n)
		}
		for i, expected := range []string{"TAG1", "TAG2", "TAG3"} {
			tag := comp[0].Input.CompletedParts[i].ETag
			if expected != *tag {
				t.Errorf("Part %d must have ETag: %s, actual: %s", i, expected, *tag)
			}
		}
	})
}

func newMockAPI(buf *bytes.Buffer) *mock_s3iot.MockS3API {
	var etag int32
	genETag := func() *string {
		i := atomic.AddInt32(&etag, 1)
		s := fmt.Sprintf("TAG%d", i)
		return &s
	}
	uploadID := "UPLOAD0"

	return &mock_s3iot.MockS3API{
		AbortMultipartUploadFunc: func(ctx context.Context, input *s3iot.AbortMultipartUploadInput) (*s3iot.AbortMultipartUploadOutput, error) {
			return &s3iot.AbortMultipartUploadOutput{}, nil
		},
		CompleteMultipartUploadFunc: func(ctx context.Context, input *s3iot.CompleteMultipartUploadInput) (*s3iot.CompleteMultipartUploadOutput, error) {
			return &s3iot.CompleteMultipartUploadOutput{
				ETag: genETag(),
			}, nil
		},
		CreateMultipartUploadFunc: func(ctx context.Context, input *s3iot.CreateMultipartUploadInput) (*s3iot.CreateMultipartUploadOutput, error) {
			return &s3iot.CreateMultipartUploadOutput{
				UploadID: &uploadID,
			}, nil
		},
		PutObjectFunc: func(ctx context.Context, input *s3iot.PutObjectInput) (*s3iot.PutObjectOutput, error) {
			io.Copy(buf, input.Body)
			return &s3iot.PutObjectOutput{
				ETag: genETag(),
			}, nil
		},
		UploadPartFunc: func(ctx context.Context, input *s3iot.UploadPartInput) (*s3iot.UploadPartOutput, error) {
			io.Copy(buf, input.Body)
			return &s3iot.UploadPartOutput{
				ETag: genETag(),
			}, nil
		},
	}
}
