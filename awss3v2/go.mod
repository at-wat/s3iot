module github.com/at-wat/s3iot/awss3v2

go 1.16

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2 v1.17.1
	github.com/aws/aws-sdk-go-v2/config v1.17.10
	github.com/aws/aws-sdk-go-v2/credentials v1.12.23
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.37
	github.com/aws/aws-sdk-go-v2/service/s3 v1.29.1
	github.com/aws/smithy-go v1.13.4
	github.com/matryer/moq v0.2.7
)
