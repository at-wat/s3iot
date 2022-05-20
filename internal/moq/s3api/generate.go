package mock_s3api // revive:disable-line:var-naming

//go:generate go run github.com/matryer/moq -pkg mock_s3api -out generated.go ../../../s3api S3API:MockS3API
