package main

import (
	"context"
	"log"
	"os"
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

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	uploader := awss3v1.NewUploader(sess)
	uc, err := uploader.Upload(context.TODO(), &s3iot.UploadInput{
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
		case <-time.After(500 * time.Millisecond):
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
