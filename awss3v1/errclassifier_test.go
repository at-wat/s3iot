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

package awss3v1

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestErrorClassifier(t *testing.T) {
	sess := session.Must(session.NewSession(
		aws.NewConfig().
			WithCredentials(credentials.AnonymousCredentials).
			WithRegion("dummy"),
	))
	_, errConnRefused := s3.New(
		sess, aws.NewConfig().WithEndpoint("http://localhost:0").WithS3ForcePathStyle(true),
	).PutObject(&s3.PutObjectInput{
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
	_, errConnReset := s3.New(
		sess, aws.NewConfig().WithEndpoint("http://"+ln.Addr().String()).WithS3ForcePathStyle(true),
	).PutObject(&s3.PutObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	testCases := map[string]struct {
		err       error
		retryable bool
		throttle  bool
		wait      time.Duration
	}{
		"Unknown": {
			err:       errors.New("an error"),
			retryable: true,
		},
		"AWSRetryable": {
			err:       awserr.New("RequestTimeout", "dummy", nil),
			retryable: true,
		},
		"AWSNonRetryable": {
			err: awserr.New("DummyFatal", "dummy", nil),
		},
		"AWSThrottle": {
			err:       awserr.New("Throttling", "dummy", nil),
			retryable: true,
			throttle:  true,
			wait:      DefaultThrottleWait,
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

	ec := &ErrorClassifier{}

	for name, tt := range testCases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Log(name, tt.err)
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
