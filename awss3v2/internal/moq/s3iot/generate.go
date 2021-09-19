package awss3v2

import (
	"github.com/at-wat/s3iot/awss3v2"
)

//go:generate go run github.com/matryer/moq -out s3iot.go . S3API:MockS3API

// S3API for generating mock.
type S3API interface {
	awss3v2.S3API
}
