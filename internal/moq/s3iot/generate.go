package mock_s3iot // revive:disable-line:var-naming

//go:generate go run github.com/matryer/moq -pkg mock_s3iot -out generated.go ../../.. S3API:MockS3API ReadInterceptorFactory:MockReadInterceptorFactory ReadInterceptor:MockReadInterceptor
