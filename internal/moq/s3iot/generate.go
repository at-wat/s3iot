package mock_s3iot

//go:generate go run github.com/matryer/moq -pkg mocks_3iot -out generated.go ../../.. S3API:MockS3API ReadInterceptorFactory:MockReadInterceptorFactory ReadInterceptor:MockReadInterceptor
