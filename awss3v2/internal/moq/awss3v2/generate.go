package mock_awss3v2

//go:generate go run github.com/matryer/moq -pkg mock_awss3v2 -out generated.go ../../.. S3API:MockS3API
