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

import (
	"context"
	"errors"
	"time"
)

func withRetry(ctx context.Context, id int64, retryer Retryer, errClassifier ErrorClassifier, fn func() error) error {
	for {
		err := fn()
		if err != nil {
			if fe, fatal := err.(*fatalError); fatal {
				return fe.error
			}
			var re *retryableError
			if !errClassifier.IsRetryable(err) && !errors.As(err, &re) {
				return err
			}
			if wait, ok := errClassifier.IsThrottle(err); ok {
				select {
				case <-time.After(wait):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if ctx.Err() == nil && retryer.OnFail(ctx, id, err) {
				continue
			}
			return err
		}
		retryer.OnSuccess(id)
		return nil
	}
}
