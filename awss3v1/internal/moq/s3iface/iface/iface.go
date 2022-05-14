package iface

import (
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// S3API for generating mock.
type S3API interface {
	s3iface.S3API
}
