package s3iot

import (
	"io"
)

// PacketizerFactory creates Packetizer for given io.Reader.
// Packetizer will be created for each Upload() call.
type PacketizerFactory interface {
	New(io.Reader) (Packetizer, error)
}

// Packetizer splits input data stream into multiple io.ReadSeekers.
type Packetizer interface {
	Len() int64
	NextReader() (io.ReadSeeker, func(), error)
}

// RetryerFactory creates Retryer.
// Retryer will be created for each Upload() call.
type RetryerFactory interface {
	New() Retryer
}

// Retryer controls upload retrying logic.
type Retryer interface {
	OnFail(id int64, err error) bool
	OnSuccess(id int64)
}

// ReadInterceptorFactory creates ReadInterceptor.
// ReadInterceptor will be created for each Upload() call.
type ReadInterceptorFactory interface {
	New() ReadInterceptor
}

// ReadInterceptor wraps io.ReadSeeker to intercept Read() calls.
type ReadInterceptor interface {
	Reader(io.ReadSeeker) io.ReadSeeker
}

// UploadStatus represents upload status.
type UploadStatus struct {
	Size         int64
	UploadedSize int64
	UploadID     string
}

// UploadOutput represents upload result.
type UploadOutput struct {
	VersionID *string
	ETag      *string
}

// UploadContext provides access to the upload progress and the result.
type UploadContext interface {
	// Result reutrns the upload status or error.
	Status() (UploadStatus, error)
	// Result reutrns the upload result or error.
	Result() (UploadOutput, error)
	// Done returns a channel which will be closed after complete.
	Done() <-chan struct{}

	// Pause the upload.
	Pause()
	// Resume the upload.
	Resume()
}

// UploadInput represents upload destination and data.
type UploadInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	Body        io.Reader
	ContentType *string
}
