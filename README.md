# S3 Manager

A Web GUI written in Go to manage S3 buckets. This was initially a (fork)(https://github.com/cloudlena/s3manager),
but things have diverged so much that currently the only thing these two have in
common, is their names.

This project includes substantial changes to make the code more idiomatic, with
support for features like server-side pagination and search (by using official 
`s3 sdk`, instead of `minio`), a revised UI, and exposing API endpoints with
OpenAPI specification file.
