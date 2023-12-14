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

package awss3v2

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
)

func TestErrorClassifier(t *testing.T) {
	// Generate actual network errors
	config := func(endpoint string) aws.Config {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider("dummy", "dummy", ""),
			),
			config.WithEndpointResolver(aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					return aws.Endpoint{URL: "http://" + endpoint}, nil
				},
			)),
			config.WithRetryer(func() aws.Retryer {
				return aws.NopRetryer{}
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		return cfg
	}
	s3Config := func(o *s3.Options) {
		o.UsePathStyle = true
	}
	_, errConnRefused := s3.NewFromConfig(
		config("localhost:0"), s3Config,
	).PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})
	ln, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, _ := ln.Accept()
		conn.Read(make([]byte, 1))
		conn.Close()
		ln.Close()
	}()
	_, errConnReset := s3.NewFromConfig(
		config(ln.Addr().String()), s3Config,
	).PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	testCases := map[string]struct {
		err       error
		retryable bool
		throttle  bool
		waitParam time.Duration
		wait      time.Duration
	}{
		"HTTP": {
			err: &http.RequestSendError{
				Err: errors.New("dummy"),
			},
			retryable: true,
		},
		"AWSRetryable": {
			err: &smithy.GenericAPIError{
				Code:    "RequestTimeout",
				Message: "dummy",
			},
			retryable: true,
		},
		"AWSNonRetryable": {
			err: &smithy.GenericAPIError{
				Code:    "DummyFatal",
				Message: "dummy",
			},
		},
		"AWSThrottleDefault": {
			err: &smithy.GenericAPIError{
				Code:    "SlowDown",
				Message: "dummy",
			},
			retryable: true,
			throttle:  true,
			wait:      DefaultThrottleWait,
		},
		"AWSThrottle": {
			err: &smithy.GenericAPIError{
				Code:    "SlowDown",
				Message: "dummy",
			},
			retryable: true,
			throttle:  true,
			waitParam: time.Minute,
			wait:      time.Minute,
		},
		"AWSConnRefused": {
			err:       errConnRefused,
			retryable: true,
		},
		"AWSConnReset": {
			err:       errConnReset,
			retryable: true,
		},
	}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Log(name, tt.err)
			ec := &ErrorClassifier{
				ThrottleWait: tt.waitParam,
			}
			if out := ec.IsRetryable(tt.err); out != tt.retryable {
				t.Errorf("IsRetryable('%v') is expected to be %v, got %v", tt.err, tt.retryable, out)
			}
			wait, throttle := ec.IsThrottle(tt.err)
			if throttle != tt.throttle {
				t.Errorf("IsThrottle('%v') is expected to be %v, got %v", tt.err, tt.throttle, throttle)
			}
			if wait != tt.wait {
				t.Errorf("Expected wait for '%v': %v, got: %v", tt.err, tt.wait, wait)
			}
		})
	}
}
