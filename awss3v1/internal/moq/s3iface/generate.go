package mock_s3iface // revive:disable-line:var-naming

//go:generate go run github.com/matryer/moq -pkg mock_s3iface -out generated.go ./iface S3API:MockS3API
