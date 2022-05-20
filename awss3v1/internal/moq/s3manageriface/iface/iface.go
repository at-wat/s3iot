package iface

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// DownloaderAPI for generating mock.
type DownloaderAPI interface {
	s3manageriface.DownloaderAPI
}

// UploaderAPI for generating mock.
type UploaderAPI interface {
	s3manageriface.UploaderAPI
}
