module github.com/at-wat/s3iot/awss3v2

go 1.16

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.42
	github.com/aws/aws-sdk-go-v2/credentials v1.13.40
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.87
	github.com/aws/aws-sdk-go-v2/service/s3 v1.40.0
	github.com/aws/smithy-go v1.14.2
	github.com/matryer/moq v0.3.2
)
