package s3iface

import (
	"github.com/at-wat/s3iot/awss3v2"
)

//go:generate go run github.com/matryer/moq -out s3iface.go . S3API:MockS3API

type S3API interface {
	awss3v2.S3API
}
