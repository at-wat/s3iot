package s3iface

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/at-wat/s3iot/awss3v2"
)

//go:generate go run github.com/matryer/moq -out s3iface.go . S3API:MockS3API HTTPClient:MockHTTPClient

// S3API for generating mock.
type S3API interface {
	awss3v2.S3API
}

// HTTPClient for generating mock.
type HTTPClient interface {
	s3.HTTPClient
}
