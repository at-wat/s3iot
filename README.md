# s3iot

[![Go Reference](https://pkg.go.dev/badge/github.com/at-wat/s3iot.svg)](https://pkg.go.dev/github.com/at-wat/s3iot) [![ci](https://github.com/at-wat/s3iot/actions/workflows/ci.yml/badge.svg)](https://github.com/at-wat/s3iot/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/at-wat/s3iot/branch/main/graph/badge.svg?token=31CXOGP3BQ)](https://codecov.io/gh/at-wat/s3iot) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## AWS S3 Uploader/Downloader for IoT-ish applications

Package s3iot provides S3 uploader/downloader applicable for unreliable and congestible network.
Main features:

- Programmable retry
- Pause/resume
- Bandwidth control (uploader only)

## Examples

- aws-sdk-go (v1)
  - [uploader](./examples/uploadv1/)
  - [downloader](./examples/downloadv1/)
- aws-sdk-go-v2
  - [uploader](./examples/uploadv2/)
  - [downloader](./examples/downloadv2/)

## License

This package is licensed under [Apache License Version 2.0](./LICENSE).
