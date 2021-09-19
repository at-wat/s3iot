package s3iface

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//go:generate go run github.com/matryer/moq -out s3iface.go . HTTPClient:MockHTTPClient

// HTTPClient for generating mock.
type HTTPClient interface {
	s3.HTTPClient
}
