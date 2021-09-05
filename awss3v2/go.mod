module github.com/at-wat/s3iot/awss3v1

go 1.17

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.15.0
)

require (
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.7.0 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
)
