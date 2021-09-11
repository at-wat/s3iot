module github.com/at-wat/s3iot/awss3v2

go 1.16

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/config v1.8.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.15.0
	github.com/matryer/moq v0.2.3
)
