package mock_awssdkv2 // revive:disable-line:var-naming

//go:generate go run github.com/matryer/moq -pkg mock_awssdkv2 -out generated.go ../../.. S3API:MockS3API
