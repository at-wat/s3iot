module github.com/at-wat/s3iot/awss3v1

go 1.23

toolchain go1.23.2

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/aws/aws-sdk-go v1.52.3
	github.com/matryer/moq v0.5.0
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
)
