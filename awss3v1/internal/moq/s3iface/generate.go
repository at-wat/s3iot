package s3iface

import (
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

//go:generate go run github.com/matryer/moq -out s3iface.go . S3API:MockS3API

type S3API interface {
	s3iface.S3API
}
