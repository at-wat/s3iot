package s3manageriface

//go:generate go run github.com/matryer/moq -pkg s3manageriface -out generated.go ./iface DownloaderAPI:MockDownloader UploaderAPI:MockUploader
