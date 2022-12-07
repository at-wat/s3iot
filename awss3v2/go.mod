module github.com/at-wat/s3iot/awss3v2

go 1.16

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2 v1.17.2
	github.com/aws/aws-sdk-go-v2/config v1.18.4
	github.com/aws/aws-sdk-go-v2/credentials v1.13.4
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.43
	github.com/aws/aws-sdk-go-v2/service/s3 v1.29.5
	github.com/aws/smithy-go v1.13.5
	github.com/matryer/moq v0.2.7
)
