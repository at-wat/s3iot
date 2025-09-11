module github.com/at-wat/s3iot/awss3v1

go 1.25

replace github.com/at-wat/s3iot => ../

require (
	github.com/at-wat/s3iot v0.0.10
	github.com/aws/aws-sdk-go v1.55.8
	github.com/matryer/moq v0.6.0
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
)
