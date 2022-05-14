package s3iface

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type HTTPClient interface {
	s3.HTTPClient
}
