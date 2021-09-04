module github.com/at-wat/s3iot/examples

go 1.17

replace github.com/at-wat/s3iot => ../

replace github.com/at-wat/s3iot/awss3v1 => ../awss3v1

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/at-wat/s3iot/awss3v1 v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.40.37
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
