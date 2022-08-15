module github.com/at-wat/s3iot/examples

go 1.16

replace github.com/at-wat/s3iot => ../

replace github.com/at-wat/s3iot/awss3v1 => ../awss3v1

replace github.com/at-wat/s3iot/awss3v2 => ../awss3v2

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/at-wat/s3iot/awss3v1 v0.0.0-00010101000000-000000000000
	github.com/at-wat/s3iot/awss3v2 v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.44.76
	github.com/aws/aws-sdk-go-v2 v1.16.11
	github.com/aws/aws-sdk-go-v2/config v1.17.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.5
)
