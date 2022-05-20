package mock_s3manageriface // revive:disable-line:var-naming

//go:generate go run github.com/matryer/moq -pkg mock_s3manageriface -out generated.go ../../../s3manageriface Uploader:MockUploader Downloader:MockDownloader
