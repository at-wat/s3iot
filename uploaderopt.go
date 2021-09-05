package s3iot

// UploaderOption sets optional parameter to the Uploader.
type UploaderOption func(*Uploader)

// WithAPI sets S3 API.
func WithAPI(a S3API) UploaderOption {
	return func(u *Uploader) {
		u.API = a
	}
}

// WithPacketizer sets PacketizerFactory to Uploader.
func WithPacketizer(p PacketizerFactory) UploaderOption {
	return func(u *Uploader) {
		u.PacketizerFactory = p
	}
}

// WithRetryer sets RetryerFactory to Uploader.
func WithRetryer(r RetryerFactory) UploaderOption {
	return func(u *Uploader) {
		u.RetryerFactory = r
	}
}

// WithReadInterceptor sets ReadInterceptorFactory to Uploader.
func WithReadInterceptor(i ReadInterceptorFactory) UploaderOption {
	return func(u *Uploader) {
		u.ReadInterceptorFactory = i
	}
}
