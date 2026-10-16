# s3iot

[![Go Reference](https://pkg.go.dev/badge/github.com/at-wat/s3iot.svg)](https://pkg.go.dev/github.com/at-wat/s3iot) [![ci](https://github.com/at-wat/s3iot/actions/workflows/ci.yml/badge.svg)](https://github.com/at-wat/s3iot/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/at-wat/s3iot/branch/main/graph/badge.svg?token=31CXOGP3BQ)](https://codecov.io/gh/at-wat/s3iot) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## News

### October 2026: awss3v2 adapter internally uses feature/s3/transfermanager
https://github.com/aws/aws-sdk-go-v2 deprecates the previously used `feature/s3/manager`.
The adapter `awss3v2` now uses `feature/s3/transfermanager` instead of `feature/s3/manager`.
If you want to use ``feature/s3/manager`, import `github.com/at-wat/s3iot/awss3v2-s3manager` instead of `github.com/at-wat/s3iot/awss3v2`.

### November 2023: API breaking change in aws-sdk-go-v2
https://github.com/aws/aws-sdk-go-v2 made a API breaking change in v1.23.0 (service/s3 v1.42.2).
If you want to use the earlier versions, import `github.com/at-wat/s3iot/awss3v2-1.22.2` instead of `github.com/at-wat/s3iot/awss3v2`.

## AWS S3 Uploader/Downloader for IoT-ish applications

Package s3iot provides S3 uploader/downloader applicable for unreliable and congestible network.
Main features:

- Programmable retry
- Pause/resume
- Bandwidth control (uploader only)

## SDK Adapters

- `awss3v1`: aws-sdk-go (v1)
- `awss3v2-1.22.2`: aws-sdk-go-v2\<=v1.22.0 with `github.com/aws/aws-sdk-go-v2/feature/s3/manager`
- `awss3v2-s3manager`: aws-sdk-go-v2 with `github.com/aws/aws-sdk-go-v2/feature/s3/manager`
- `awss3v2`: aws-sdk-go-v2 with `github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager`

## Examples

- aws-sdk-go (v1)
  - [uploader](./examples/uploadv1/main.go)
  - [downloader](./examples/downloadv1/main.go)
- aws-sdk-go-v2
  - [uploader](./examples/uploadv2/main.go)
  - [downloader](./examples/downloadv2/main.go)

## License

This package is licensed under [Apache License Version 2.0](./LICENSE).
