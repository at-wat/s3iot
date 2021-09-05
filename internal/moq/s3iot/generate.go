package s3iface

import (
	"github.com/at-wat/s3iot"
)

//go:generate go run github.com/matryer/moq -out s3api.go . S3API:MockS3API

type S3API interface {
	s3iot.S3API
}
