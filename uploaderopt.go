// Copyright 2021 The s3iot authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3iot

// UploaderOption sets optional parameter to the Uploader.
type UploaderOption func(*Uploader)

// WithAPI sets S3 API.
func WithAPI(a S3API) UploaderOption {
	return func(u *Uploader) {
		u.API = a
	}
}

// WithUploadSlicer sets UploadSlicerFactory to Uploader.
func WithUploadSlicer(p UploadSlicerFactory) UploaderOption {
	return func(u *Uploader) {
		u.UploadSlicerFactory = p
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
