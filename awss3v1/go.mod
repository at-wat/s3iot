module github.com/at-wat/s3iot/awss3v1

go 1.18

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/aws/aws-sdk-go v1.55.8
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
