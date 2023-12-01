module github.com/at-wat/s3iot/examples

go 1.16

replace github.com/at-wat/s3iot => ../

replace github.com/at-wat/s3iot/awss3v1 => ../awss3v1

replace github.com/at-wat/s3iot/awss3v2 => ../awss3v2

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/at-wat/s3iot/awss3v1 v0.0.0-00010101000000-000000000000
	github.com/at-wat/s3iot/awss3v2 v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.48.11
	github.com/aws/aws-sdk-go-v2 v1.21.1
	github.com/aws/aws-sdk-go-v2/config v1.18.44
	github.com/aws/aws-sdk-go-v2/service/s3 v1.40.1
)
