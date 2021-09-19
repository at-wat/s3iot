package mock_s3iot

import (
	"github.com/at-wat/s3iot"
)

//go:generate go run github.com/matryer/moq -out s3iot.go . S3API:MockS3API ReadInterceptorFactory:MockReadInterceptorFactory ReadInterceptor:MockReadInterceptor

// S3API for generating mock.
type S3API interface {
	s3iot.S3API
}

// ReadInterceptorFactory for generating mock.
type ReadInterceptorFactory interface {
	s3iot.ReadInterceptorFactory
}

// ReadInterceptor for generating mock.
type ReadInterceptor interface {
	s3iot.ReadInterceptor
}
