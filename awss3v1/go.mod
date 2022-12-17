module github.com/at-wat/s3iot/awss3v1

go 1.19

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.44.154
	github.com/matryer/moq v0.2.7
)
