package s3iot

import (
	"context"
	"io"
	"sort"
	"sync"
)

type completedParts []*CompletedPart

func (a completedParts) Len() int {
	return len(a)
}

func (a completedParts) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a completedParts) Less(i, j int) bool {
	return *a[i].PartNumber < *a[j].PartNumber
}

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

type Uploader struct {
	API               S3API
	PacketizerFactory PacketizerFactory
	RetryerFactory    RetryerFactory
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

func (u Uploader) Upload(ctx context.Context, input *UploadInput) (UploadContext, error) {
	if u.PacketizerFactory == nil {
		u.PacketizerFactory = &DefaultPacketizerFactory{}
	}
	if u.RetryerFactory == nil {
		u.RetryerFactory = DefaultRetryer
	}
	packetizer, err := u.PacketizerFactory.New(input.Body)
	if err != nil {
		return nil, err
	}
	uc := &uploadContext{
		api:        u.API,
		retryer:    u.RetryerFactory.New(),
		packetizer: packetizer,
		input:      input,
		done:       make(chan struct{}),
		paused:     make(chan struct{}),
		status: UploadStatus{
			Size: packetizer.Len(),
		},
	}
	close(uc.paused)
	uc.resumeOnce.Do(func() {})
	r, cleanup, err := uc.packetizer.NextReader()
	if err == io.EOF {
		go uc.single(ctx, r, cleanup)
		return uc, nil
	} else if err != nil {
		return nil, err
	}
	go uc.multi(ctx, r, cleanup)
	return uc, nil
}

type uploadContext struct {
	api        S3API
	packetizer Packetizer
	retryer    Retryer
	input      *UploadInput

	status UploadStatus
	output UploadOutput
	err    error

	paused     chan struct{}
	resumeOnce sync.Once

	mu   sync.RWMutex
	done chan struct{}
}

func (uc *uploadContext) Done() <-chan struct{} {
	return uc.done
}

func (uc *uploadContext) Pause() {
	uc.mu.Lock()
	uc.paused = make(chan struct{})
	uc.resumeOnce = sync.Once{}
	uc.mu.Unlock()
}

func (uc *uploadContext) Resume() {
	uc.mu.RLock()
	uc.resumeOnce.Do(func() {
		close(uc.paused)
	})
	uc.mu.RUnlock()
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

	if err := withRetry(ctx, 0, uc.retryer, func() error {
		out, err := uc.api.PutObject(ctx, &PutObjectInput{
			Bucket:      uc.input.Bucket,
			Key:         uc.input.Key,
			ACL:         uc.input.ACL,
			Body:        r,
			ContentType: uc.input.ContentType,
		})
		if err != nil {
			return err
		}
		uc.success(UploadOutput{
			out.VersionID,
			out.ETag,
		})
		return nil
	}); err != nil {
		uc.fail(err)
	}
}

func (uc *uploadContext) multi(ctx context.Context, r io.ReadSeeker, cleanup func()) {
	if err := withRetry(ctx, 0, uc.retryer, func() error {
		out, err := uc.api.CreateMultipartUpload(ctx, &CreateMultipartUploadInput{
			Bucket:      uc.input.Bucket,
			Key:         uc.input.Key,
			ACL:         uc.input.ACL,
			ContentType: uc.input.ContentType,
		})
		if err != nil {
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
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			cleanup()
			uc.fail(err)
			return
		}
		if err := withRetry(ctx, i, uc.retryer, func() error {
			out, err := uc.api.UploadPart(ctx, &UploadPartInput{
				Body:       r,
				Bucket:     uc.input.Bucket,
				Key:        uc.input.Key,
				PartNumber: &i,
				UploadID:   &uc.status.UploadID,
			})
			if err != nil {
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
		uc.status.UploadedSize += size
		paused := uc.paused
		uc.mu.Unlock()

		if last {
			break
		}
		select {
		case <-paused:
		case <-ctx.Done():
			uc.fail(ctx.Err())
			return
		}

		r, cleanup, err = uc.packetizer.NextReader()
		switch {
		case err == io.EOF:
			last = true
		case err != nil:
			uc.fail(err)
			return
		}
	}
	sort.Sort(parts)

	if err := withRetry(ctx, -1, uc.retryer, func() error {
		out, err := uc.api.CompleteMultipartUpload(ctx, &CompleteMultipartUploadInput{
			Bucket:         uc.input.Bucket,
			Key:            uc.input.Key,
			CompletedParts: parts,
			UploadID:       &uc.status.UploadID,
		})
		if err != nil {
			return err
		}
		uc.success(UploadOutput{
			out.VersionID,
			out.ETag,
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
