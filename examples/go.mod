module github.com/at-wat/s3iot/examples

go 1.17

replace github.com/at-wat/s3iot => ../

replace github.com/at-wat/s3iot/awss3v1 => ../awss3v1

replace github.com/at-wat/s3iot/awss3v2 => ../awss3v2

require (
	github.com/at-wat/s3iot v0.0.0-00010101000000-000000000000
	github.com/at-wat/s3iot/awss3v1 v0.0.0-00010101000000-000000000000
	github.com/at-wat/s3iot/awss3v2 v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.40.37
	github.com/aws/aws-sdk-go-v2 v1.9.0
	github.com/aws/aws-sdk-go-v2/config v1.8.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.15.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.7.0 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)
