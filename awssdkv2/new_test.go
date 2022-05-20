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

package awssdkv2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/at-wat/s3iot"
)

func TestNew(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("NewUploader", func(t *testing.T) {
		u := NewUploader(cfg, s3iot.UpDownloaderOptionFn(func(u *s3iot.UpDownloaderBase) {
			if _, ok := u.API.(*wrapper).api.(*s3.Client); !ok {
				t.Errorf("Base API is expected to be *s3.Client, actually %T", u.API.(*wrapper).api)
			}
		}))
		if _, ok := u.API.(*wrapper).api.(*s3.Client); !ok {
			t.Errorf("Base API is expected to be *s3.Client, actually %T", u.API.(*wrapper).api)
		}
	})
	t.Run("NewDownloader", func(t *testing.T) {
		d := NewDownloader(cfg, s3iot.UpDownloaderOptionFn(func(u *s3iot.UpDownloaderBase) {
			if _, ok := u.API.(*wrapper).api.(*s3.Client); !ok {
				t.Errorf("Base API is expected to be *s3.Client, actually %T", u.API.(*wrapper).api)
			}
		}))
		if _, ok := d.API.(*wrapper).api.(*s3.Client); !ok {
			t.Errorf("Base API is expected to be *s3.Client, actually %T", d.API.(*wrapper).api)
		}
	})
}
