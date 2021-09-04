package s3iot

import (
	"context"
)

type NoRetryerFactory struct{}

func (NoRetryerFactory) New() Retryer {
	return &noRetryer{}
}

type noRetryer struct{}

func (noRetryer) OnFail(error) bool {
	return false
}

func withRetry(ctx context.Context, retryer Retryer, fn func() error) error {
	for {
		err := fn()
		if err != nil {
			if ctx.Err() == nil && retryer.OnFail(err) {
				continue
			}
			return err
		}
		return nil
	}
}
