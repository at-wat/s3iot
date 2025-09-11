package mock_awss3v2 // revive:disable-line:var-naming

// Since moq declares only the latest Go version is supported,
// moq version can't be controlled under our go.mod.
//
// So, moq must be installed by:
//   cd s3iot/tools
//   go install github.com/matryer/moq

//go:generate moq -pkg mock_awss3v2 -out generated.go ../../.. S3API:MockS3API
