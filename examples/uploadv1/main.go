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

package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/awss3v1"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s file bucket key", os.Args[0])
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession(&aws.Config{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          10,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				WriteBufferSize:       1024, // Reduce bandwidth burst
			},
			Timeout: time.Minute,
		},
		MaxRetries: aws.Int(0), // Use retry logic in s3iot
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		cancel()
	}()

	uploader := awss3v1.NewUploader(sess,
		s3iot.WithReadInterceptor(
			s3iot.NewWaitReadInterceptorFactory(time.Microsecond), // 1s/MB
		),
	)
	uc, err := uploader.Upload(ctx, &s3iot.UploadInput{
		Bucket: aws.String(os.Args[2]),
		Key:    aws.String(os.Args[3]),
		Body:   f,
	})
	if err != nil {
		log.Fatal(err)
	}

	showStatus := func() {
		status, err := uc.Status()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%+v", status)
	}

	for {
		select {
		case <-time.After(time.Second):
			showStatus()
		case <-uc.Done():
			showStatus()
			out, err := uc.Result()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%+v", out)
			return
		}
	}
}
