package s3manageriface

// Since moq declares only the latest Go version is supported,
// moq version can't be controlled under our go.mod.
//
// So, moq must be installed by:
//   cd s3iot/tools
//   go install github.com/matryer/moq

//go:generate moq -pkg s3manageriface -out generated.go ./iface DownloaderAPI:MockDownloader UploaderAPI:MockUploader
