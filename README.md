# S3 Manager

Backend and Web GUI for managing S3 compatible buckets.  
This was initially a [fork](https://github.com/cloudlena/s3manager), but things
have diverged so much that currently the only thing these two have in common, is
their names.

| ![](./assets/images/buckets.png)<br/>Buckets | ![](./assets/images/objects.png)<br/>Objects |
|----------------------------------------------|----------------------------------------------|

This project includes substantial changes to make the code more idiomatic, with
support for features like server-side pagination and search (by using official 
`aws-sdk`, instead of `minio`), a revised UI, and exposing API endpoints with
OpenAPI specification file.
