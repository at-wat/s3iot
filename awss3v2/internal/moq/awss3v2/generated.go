// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock_awss3v2

import (
	"context"
	"github.com/at-wat/s3iot/awss3v2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sync"
)

// Ensure, that MockS3API does implement awss3v2.S3API.
// If this is not the case, regenerate this file with moq.
var _ awss3v2.S3API = &MockS3API{}

// MockS3API is a mock implementation of awss3v2.S3API.
//
// 	func TestSomethingThatUsesS3API(t *testing.T) {
//
// 		// make and configure a mocked awss3v2.S3API
// 		mockedS3API := &MockS3API{
// 			AbortMultipartUploadFunc: func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
// 				panic("mock out the AbortMultipartUpload method")
// 			},
// 			CompleteMultipartUploadFunc: func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
// 				panic("mock out the CompleteMultipartUpload method")
// 			},
// 			CreateMultipartUploadFunc: func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
// 				panic("mock out the CreateMultipartUpload method")
// 			},
// 			DeleteObjectFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
// 				panic("mock out the DeleteObject method")
// 			},
// 			GetObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
// 				panic("mock out the GetObject method")
// 			},
// 			ListObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
// 				panic("mock out the ListObjectsV2 method")
// 			},
// 			PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
// 				panic("mock out the PutObject method")
// 			},
// 			UploadPartFunc: func(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
// 				panic("mock out the UploadPart method")
// 			},
// 		}
//
// 		// use mockedS3API in code that requires awss3v2.S3API
// 		// and then make assertions.
//
// 	}
type MockS3API struct {
	// AbortMultipartUploadFunc mocks the AbortMultipartUpload method.
	AbortMultipartUploadFunc func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error)

	// CompleteMultipartUploadFunc mocks the CompleteMultipartUpload method.
	CompleteMultipartUploadFunc func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error)

	// CreateMultipartUploadFunc mocks the CreateMultipartUpload method.
	CreateMultipartUploadFunc func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error)

	// DeleteObjectFunc mocks the DeleteObject method.
	DeleteObjectFunc func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)

	// GetObjectFunc mocks the GetObject method.
	GetObjectFunc func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)

	// ListObjectsV2Func mocks the ListObjectsV2 method.
	ListObjectsV2Func func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)

	// PutObjectFunc mocks the PutObject method.
	PutObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

	// UploadPartFunc mocks the UploadPart method.
	UploadPartFunc func(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error)

	// calls tracks calls to the methods.
	calls struct {
		// AbortMultipartUpload holds details about calls to the AbortMultipartUpload method.
		AbortMultipartUpload []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.AbortMultipartUploadInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// CompleteMultipartUpload holds details about calls to the CompleteMultipartUpload method.
		CompleteMultipartUpload []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.CompleteMultipartUploadInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// CreateMultipartUpload holds details about calls to the CreateMultipartUpload method.
		CreateMultipartUpload []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.CreateMultipartUploadInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// DeleteObject holds details about calls to the DeleteObject method.
		DeleteObject []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.DeleteObjectInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// GetObject holds details about calls to the GetObject method.
		GetObject []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.GetObjectInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// ListObjectsV2 holds details about calls to the ListObjectsV2 method.
		ListObjectsV2 []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.ListObjectsV2Input
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// PutObject holds details about calls to the PutObject method.
		PutObject []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.PutObjectInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
		// UploadPart holds details about calls to the UploadPart method.
		UploadPart []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *s3.UploadPartInput
			// OptFns is the optFns argument value.
			OptFns []func(*s3.Options)
		}
	}
	lockAbortMultipartUpload    sync.RWMutex
	lockCompleteMultipartUpload sync.RWMutex
	lockCreateMultipartUpload   sync.RWMutex
	lockDeleteObject            sync.RWMutex
	lockGetObject               sync.RWMutex
	lockListObjectsV2           sync.RWMutex
	lockPutObject               sync.RWMutex
	lockUploadPart              sync.RWMutex
}

// AbortMultipartUpload calls AbortMultipartUploadFunc.
func (mock *MockS3API) AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
	if mock.AbortMultipartUploadFunc == nil {
		panic("MockS3API.AbortMultipartUploadFunc: method is nil but S3API.AbortMultipartUpload was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.AbortMultipartUploadInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockAbortMultipartUpload.Lock()
	mock.calls.AbortMultipartUpload = append(mock.calls.AbortMultipartUpload, callInfo)
	mock.lockAbortMultipartUpload.Unlock()
	return mock.AbortMultipartUploadFunc(ctx, params, optFns...)
}

// AbortMultipartUploadCalls gets all the calls that were made to AbortMultipartUpload.
// Check the length with:
//     len(mockedS3API.AbortMultipartUploadCalls())
func (mock *MockS3API) AbortMultipartUploadCalls() []struct {
	Ctx    context.Context
	Params *s3.AbortMultipartUploadInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.AbortMultipartUploadInput
		OptFns []func(*s3.Options)
	}
	mock.lockAbortMultipartUpload.RLock()
	calls = mock.calls.AbortMultipartUpload
	mock.lockAbortMultipartUpload.RUnlock()
	return calls
}

// CompleteMultipartUpload calls CompleteMultipartUploadFunc.
func (mock *MockS3API) CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	if mock.CompleteMultipartUploadFunc == nil {
		panic("MockS3API.CompleteMultipartUploadFunc: method is nil but S3API.CompleteMultipartUpload was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.CompleteMultipartUploadInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockCompleteMultipartUpload.Lock()
	mock.calls.CompleteMultipartUpload = append(mock.calls.CompleteMultipartUpload, callInfo)
	mock.lockCompleteMultipartUpload.Unlock()
	return mock.CompleteMultipartUploadFunc(ctx, params, optFns...)
}

// CompleteMultipartUploadCalls gets all the calls that were made to CompleteMultipartUpload.
// Check the length with:
//     len(mockedS3API.CompleteMultipartUploadCalls())
func (mock *MockS3API) CompleteMultipartUploadCalls() []struct {
	Ctx    context.Context
	Params *s3.CompleteMultipartUploadInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.CompleteMultipartUploadInput
		OptFns []func(*s3.Options)
	}
	mock.lockCompleteMultipartUpload.RLock()
	calls = mock.calls.CompleteMultipartUpload
	mock.lockCompleteMultipartUpload.RUnlock()
	return calls
}

// CreateMultipartUpload calls CreateMultipartUploadFunc.
func (mock *MockS3API) CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	if mock.CreateMultipartUploadFunc == nil {
		panic("MockS3API.CreateMultipartUploadFunc: method is nil but S3API.CreateMultipartUpload was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.CreateMultipartUploadInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockCreateMultipartUpload.Lock()
	mock.calls.CreateMultipartUpload = append(mock.calls.CreateMultipartUpload, callInfo)
	mock.lockCreateMultipartUpload.Unlock()
	return mock.CreateMultipartUploadFunc(ctx, params, optFns...)
}

// CreateMultipartUploadCalls gets all the calls that were made to CreateMultipartUpload.
// Check the length with:
//     len(mockedS3API.CreateMultipartUploadCalls())
func (mock *MockS3API) CreateMultipartUploadCalls() []struct {
	Ctx    context.Context
	Params *s3.CreateMultipartUploadInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.CreateMultipartUploadInput
		OptFns []func(*s3.Options)
	}
	mock.lockCreateMultipartUpload.RLock()
	calls = mock.calls.CreateMultipartUpload
	mock.lockCreateMultipartUpload.RUnlock()
	return calls
}

// DeleteObject calls DeleteObjectFunc.
func (mock *MockS3API) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if mock.DeleteObjectFunc == nil {
		panic("MockS3API.DeleteObjectFunc: method is nil but S3API.DeleteObject was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.DeleteObjectInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockDeleteObject.Lock()
	mock.calls.DeleteObject = append(mock.calls.DeleteObject, callInfo)
	mock.lockDeleteObject.Unlock()
	return mock.DeleteObjectFunc(ctx, params, optFns...)
}

// DeleteObjectCalls gets all the calls that were made to DeleteObject.
// Check the length with:
//     len(mockedS3API.DeleteObjectCalls())
func (mock *MockS3API) DeleteObjectCalls() []struct {
	Ctx    context.Context
	Params *s3.DeleteObjectInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.DeleteObjectInput
		OptFns []func(*s3.Options)
	}
	mock.lockDeleteObject.RLock()
	calls = mock.calls.DeleteObject
	mock.lockDeleteObject.RUnlock()
	return calls
}

// GetObject calls GetObjectFunc.
func (mock *MockS3API) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if mock.GetObjectFunc == nil {
		panic("MockS3API.GetObjectFunc: method is nil but S3API.GetObject was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.GetObjectInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockGetObject.Lock()
	mock.calls.GetObject = append(mock.calls.GetObject, callInfo)
	mock.lockGetObject.Unlock()
	return mock.GetObjectFunc(ctx, params, optFns...)
}

// GetObjectCalls gets all the calls that were made to GetObject.
// Check the length with:
//     len(mockedS3API.GetObjectCalls())
func (mock *MockS3API) GetObjectCalls() []struct {
	Ctx    context.Context
	Params *s3.GetObjectInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.GetObjectInput
		OptFns []func(*s3.Options)
	}
	mock.lockGetObject.RLock()
	calls = mock.calls.GetObject
	mock.lockGetObject.RUnlock()
	return calls
}

// ListObjectsV2 calls ListObjectsV2Func.
func (mock *MockS3API) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if mock.ListObjectsV2Func == nil {
		panic("MockS3API.ListObjectsV2Func: method is nil but S3API.ListObjectsV2 was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.ListObjectsV2Input
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockListObjectsV2.Lock()
	mock.calls.ListObjectsV2 = append(mock.calls.ListObjectsV2, callInfo)
	mock.lockListObjectsV2.Unlock()
	return mock.ListObjectsV2Func(ctx, params, optFns...)
}

// ListObjectsV2Calls gets all the calls that were made to ListObjectsV2.
// Check the length with:
//     len(mockedS3API.ListObjectsV2Calls())
func (mock *MockS3API) ListObjectsV2Calls() []struct {
	Ctx    context.Context
	Params *s3.ListObjectsV2Input
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.ListObjectsV2Input
		OptFns []func(*s3.Options)
	}
	mock.lockListObjectsV2.RLock()
	calls = mock.calls.ListObjectsV2
	mock.lockListObjectsV2.RUnlock()
	return calls
}

// PutObject calls PutObjectFunc.
func (mock *MockS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if mock.PutObjectFunc == nil {
		panic("MockS3API.PutObjectFunc: method is nil but S3API.PutObject was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.PutObjectInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockPutObject.Lock()
	mock.calls.PutObject = append(mock.calls.PutObject, callInfo)
	mock.lockPutObject.Unlock()
	return mock.PutObjectFunc(ctx, params, optFns...)
}

// PutObjectCalls gets all the calls that were made to PutObject.
// Check the length with:
//     len(mockedS3API.PutObjectCalls())
func (mock *MockS3API) PutObjectCalls() []struct {
	Ctx    context.Context
	Params *s3.PutObjectInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.PutObjectInput
		OptFns []func(*s3.Options)
	}
	mock.lockPutObject.RLock()
	calls = mock.calls.PutObject
	mock.lockPutObject.RUnlock()
	return calls
}

// UploadPart calls UploadPartFunc.
func (mock *MockS3API) UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	if mock.UploadPartFunc == nil {
		panic("MockS3API.UploadPartFunc: method is nil but S3API.UploadPart was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params *s3.UploadPartInput
		OptFns []func(*s3.Options)
	}{
		Ctx:    ctx,
		Params: params,
		OptFns: optFns,
	}
	mock.lockUploadPart.Lock()
	mock.calls.UploadPart = append(mock.calls.UploadPart, callInfo)
	mock.lockUploadPart.Unlock()
	return mock.UploadPartFunc(ctx, params, optFns...)
}

// UploadPartCalls gets all the calls that were made to UploadPart.
// Check the length with:
//     len(mockedS3API.UploadPartCalls())
func (mock *MockS3API) UploadPartCalls() []struct {
	Ctx    context.Context
	Params *s3.UploadPartInput
	OptFns []func(*s3.Options)
} {
	var calls []struct {
		Ctx    context.Context
		Params *s3.UploadPartInput
		OptFns []func(*s3.Options)
	}
	mock.lockUploadPart.RLock()
	calls = mock.calls.UploadPart
	mock.lockUploadPart.RUnlock()
	return calls
}
