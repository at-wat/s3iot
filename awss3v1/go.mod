module github.com/at-wat/s3iot/awss3v1

go 1.16

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/aws/aws-sdk-go v1.44.172
	github.com/matryer/moq v0.3.0
)
