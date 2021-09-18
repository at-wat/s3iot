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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/aws/aws-sdk-go-v2/service/s3" // lock version as direct dependency

	"github.com/at-wat/s3iot"
	"github.com/at-wat/s3iot/awss3v2"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s file bucket key", os.Args[0])
	}

	f, err := os.Create(os.Args[1])
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

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          10,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: time.Minute,
		}),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.Retryables = []retry.IsErrorRetryable{
					retry.IsErrorRetryableFunc(func(err error) aws.Ternary { return aws.FalseTernary }),
				}
			})
		}), // Use retry logic in s3iot
	)
	if err != nil {
		log.Fatal(err)
	}

	uploader := awss3v2.NewDownloader(cfg,
		s3iot.WithRetryer(&s3iot.RetryerHookFactory{
			Base: s3iot.DefaultRetryer,
			OnError: func(bucket, key string, err error) {
				log.Print(bucket, key, err)
			},
		}),
	)
	dc, err := uploader.Download(ctx, f, &s3iot.DownloadInput{
		Bucket: aws.String(os.Args[2]),
		Key:    aws.String(os.Args[3]),
	})
	if err != nil {
		log.Fatal(err)
	}

	showStatus := func() {
		status, err := dc.Status()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%+v", status)
	}

	for {
		select {
		case <-time.After(time.Second):
			showStatus()
		case <-dc.Done():
			showStatus()
			out, err := dc.Result()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%+v", out)
			return
		}
	}
}
