# S3 Manager

A Web GUI written in Go to manage S3 buckets. This was initially a fork of this
[project](github.com/cloudlena/s3manager), but things have diverged so much that
currently the only thing these two have in common, is their names.

This project includes substantial changes to make the code more idiomatic, with
support for features like pagination and search, a freshly revised UI, replacing
the `minio` with official `Amazon s3` sdk, and exposing API endpoints with
accompanying OpenAPI specification file.
