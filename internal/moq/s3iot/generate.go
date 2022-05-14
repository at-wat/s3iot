package s3iot

//go:generate go run github.com/matryer/moq -out s3iot.go ../../.. S3API:MockS3API ReadInterceptorFactory:MockReadInterceptorFactory ReadInterceptor:MockReadInterceptor
