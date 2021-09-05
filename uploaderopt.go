package s3iot

type UploaderOption func(*Uploader)

func WithAPI(a S3API) UploaderOption {
	return func(u *Uploader) {
		u.API = a
	}
}

func WithPacketizer(p PacketizerFactory) UploaderOption {
	return func(u *Uploader) {
		u.PacketizerFactory = p
	}
}

func WithRetryer(r RetryerFactory) UploaderOption {
	return func(u *Uploader) {
		u.RetryerFactory = r
	}
}

func WithReadInterceptor(i ReadInterceptorFactory) UploaderOption {
	return func(u *Uploader) {
		u.ReadInterceptorFactory = i
	}
}
