package s3iface

//go:generate go run github.com/matryer/moq -pkg s3iface -out generated.go ./iface S3API:MockS3API
