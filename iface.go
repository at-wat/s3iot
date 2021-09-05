package s3iot

import (
	"io"
)

type PacketizerFactory interface {
	New(io.Reader) (Packetizer, error)
}

type Packetizer interface {
	Len() int64
	NextReader() (io.ReadSeeker, func(), error)
}

type RetryerFactory interface {
	New() Retryer
}

type Retryer interface {
	OnFail(id int64, err error) bool
	OnSuccess(id int64)
}

type UploadStatus struct {
	Size         int64
	UploadedSize int64
	UploadID     string
}

type UploadOutput struct {
	VersionID *string
	ETag      *string
}

type UploadContext interface {
	Status() (UploadStatus, error)
	Result() (UploadOutput, error)
	Done() <-chan struct{}

	Pause()
	Resume()
}

type UploadInput struct {
	Bucket      *string
	Key         *string
	ACL         *string
	Body        io.Reader
	ContentType *string
}
