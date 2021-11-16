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

// Package s3iot provides S3 uploader applicable for unreliable and congestible network.
// Object can be uploaded with retry, pause/resume, and bandwidth limit.
package s3iot

import (
	"context"
	"io"
	"sort"
)

// Upload a file to S3.
func (u Uploader) Upload(ctx context.Context, input *UploadInput) (UploadContext, error) {
	if u.UploadSlicerFactory == nil {
		u.UploadSlicerFactory = &DefaultUploadSlicerFactory{}
	}
	if u.RetryerFactory == nil {
		u.RetryerFactory = DefaultRetryer
	}
	if u.ErrorClassifier == nil {
		u.ErrorClassifier = DefaultErrorClassifier
	}
	slicer, err := u.UploadSlicerFactory.New(input.Body)
	if err != nil {
		return nil, err
	}
	var readInterceptor ReadInterceptor
	if u.ReadInterceptorFactory != nil {
		readInterceptor = u.ReadInterceptorFactory.New()
	}
	uc := &uploadContext{
		upDownloadContext: newUpDownloadContext(
			u.API,
			u.RetryerFactory,
			u.ErrorClassifier,
			u.ForcePause,
		),
		slicer:          slicer,
		readInterceptor: readInterceptor,
		input:           input,
		status: UploadStatus{
			Status: Status{
				Size: slicer.Len(),
			},
		},
	}
	uc.setStatePtr(&uc.status.Paused, &uc.status.NumRetries)
	r, cleanup, err := uc.slicer.NextReader()
	switch {
	case err == io.EOF:
		go uc.single(ctx, r, cleanup)
		return uc, nil
	case err != nil:
		return nil, err
	}
	go uc.multi(ctx, r, cleanup)
	return uc, nil
}

type uploadContext struct {
	*upDownloadContext

	slicer          UploadSlicer
	readInterceptor ReadInterceptor
	input           *UploadInput

	status UploadStatus
	output UploadOutput
}

func (uc *uploadContext) BucketKey() (bucket, key string) {
	return *uc.input.Bucket, *uc.input.Key
}

func (uc *uploadContext) Status() (UploadStatus, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.status, uc.err
}

func (uc *uploadContext) Result() (UploadOutput, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.output, uc.err
}

func (uc *uploadContext) single(ctx context.Context, r io.ReadSeeker, cleanup func()) {
	defer cleanup()

	if uc.readInterceptor != nil {
		r = uc.readInterceptor.Reader(r)
	}

	if err := withRetry(ctx, 0, uc.retryer, uc.errClassifier, func() error {
		uc.pauseCheck(ctx)
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			return &fatalError{err}
		}
		ctx2, cancel := uc.currentCallContext(ctx)
		out, err := uc.api.PutObject(ctx2, &PutObjectInput{
			Bucket:      uc.input.Bucket,
			Key:         uc.input.Key,
			ACL:         uc.input.ACL,
			Body:        r,
			ContentType: uc.input.ContentType,
		})
		cancel()
		if err != nil {
			uc.countRetry()
			return err
		}
		uc.success(UploadOutput{
			VersionID: out.VersionID,
			ETag:      out.ETag,
			Location:  out.Location,
		})
		return nil
	}); err != nil {
		uc.fail(err)
	}
}

func (uc *uploadContext) multi(ctx context.Context, r io.ReadSeeker, cleanup func()) {
	if err := withRetry(ctx, 0, uc.retryer, uc.errClassifier, func() error {
		uc.pauseCheck(ctx)
		out, err := uc.api.CreateMultipartUpload(ctx, &CreateMultipartUploadInput{
			Bucket:      uc.input.Bucket,
			Key:         uc.input.Key,
			ACL:         uc.input.ACL,
			ContentType: uc.input.ContentType,
		})
		if err != nil {
			uc.countRetry()
			return err
		}
		uc.mu.Lock()
		uc.status.UploadID = *out.UploadID
		uc.mu.Unlock()
		return nil
	}); err != nil {
		cleanup()
		uc.fail(err)
		return
	}

	var parts completedParts
	var last bool
	for i := int64(1); ; i++ {
		i := i
		size, err := r.Seek(0, io.SeekEnd)
		if err != nil {
			cleanup()
			uc.fail(err)
			return
		}
		if size == 0 {
			break
		}
		if uc.readInterceptor != nil {
			r = uc.readInterceptor.Reader(r)
		}
		if err := withRetry(ctx, i, uc.retryer, uc.errClassifier, func() error {
			uc.pauseCheck(ctx)
			if _, err := r.Seek(0, io.SeekStart); err != nil {
				return &fatalError{err}
			}
			ctx2, cancel := uc.currentCallContext(ctx)
			out, err := uc.api.UploadPart(ctx2, &UploadPartInput{
				Body:       r,
				Bucket:     uc.input.Bucket,
				Key:        uc.input.Key,
				PartNumber: &i,
				UploadID:   &uc.status.UploadID,
			})
			cancel()
			if err != nil {
				uc.countRetry()
				return err
			}
			parts = append(parts, &CompletedPart{
				PartNumber: &i,
				ETag:       out.ETag,
			})
			return nil
		}); err != nil {
			cleanup()
			uc.fail(err)
			return
		}
		cleanup()
		uc.mu.Lock()
		uc.status.CompletedSize += size
		uc.mu.Unlock()

		if last {
			break
		}

		r, cleanup, err = uc.slicer.NextReader()
		switch {
		case err == io.EOF:
			last = true
		case err != nil:
			uc.fail(err)
			return
		}
	}
	sort.Sort(parts)

	if err := withRetry(ctx, -1, uc.retryer, uc.errClassifier, func() error {
		uc.pauseCheck(ctx)
		out, err := uc.api.CompleteMultipartUpload(ctx, &CompleteMultipartUploadInput{
			Bucket:         uc.input.Bucket,
			Key:            uc.input.Key,
			CompletedParts: parts,
			UploadID:       &uc.status.UploadID,
		})
		if err != nil {
			uc.countRetry()
			return err
		}
		uc.success(UploadOutput{
			VersionID: out.VersionID,
			ETag:      out.ETag,
			Location:  out.Location,
		})
		return nil
	}); err != nil {
		uc.fail(err)
	}
}

func (uc *uploadContext) fail(err error) {
	uc.mu.Lock()
	uc.err = err
	uc.mu.Unlock()
	close(uc.done)

	_, _ = uc.api.AbortMultipartUpload(context.Background(), &AbortMultipartUploadInput{
		Bucket:   uc.input.Bucket,
		Key:      uc.input.Key,
		UploadID: &uc.status.UploadID,
	})
}

func (uc *uploadContext) success(out UploadOutput) {
	uc.mu.Lock()
	uc.output = out
	uc.mu.Unlock()
	close(uc.done)
}
