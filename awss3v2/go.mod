module github.com/at-wat/s3iot/awss3v2

go 1.16

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/aws/aws-sdk-go-v2 v1.30.1
	github.com/aws/aws-sdk-go-v2/config v1.27.23
	github.com/aws/aws-sdk-go-v2/credentials v1.17.23
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.57.1
	github.com/aws/smithy-go v1.20.3
	github.com/matryer/moq v0.3.4
)
