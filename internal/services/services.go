package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"github.com/hossein1376/s3manager/internal/model"
	"github.com/hossein1376/s3manager/pkg/errs"
)

var (
	ErrBucketNotEmpty = errors.New("bucket is not empty")
	ErrExistingBucket = errors.New("bucket already exists")
	ErrMissingBucket  = errors.New("bucket not found")
)

type Services struct {
	s3Client *s3.Client
}

func New(s3Client *s3.Client) *Services {
	return &Services{s3Client: s3Client}
}

func (s *Services) ListObjects(
	ctx context.Context,
	bucketName string,
	maxKeys int32,
	opt model.ListObjectsOption,
) ([]model.Object, *string, error) {
	params := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucketName),
		MaxKeys:           aws.Int32(maxKeys),
		ContinuationToken: opt.ContinuationToken,
		Prefix:            opt.Filter,
	}
	list, err := s.s3Client.ListObjectsV2(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchBucket") {
			return nil, nil, errs.NotFound(errs.WithErr(ErrMissingBucket))
		}
		return nil, nil, err
	}
	objects := make([]model.Object, *list.KeyCount)
	for i, obj := range list.Contents {
		objects[i] = model.Object{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		}
	}
	return objects, list.NextContinuationToken, nil
}

func (s *Services) ListBuckets(
	ctx context.Context, count int32, opts model.ListBucketsOptions,
) ([]model.Bucket, *string, error) {
	params := &s3.ListBucketsInput{
		ContinuationToken: opts.ContinuationToken,
		MaxBuckets:        aws.Int32(count),
		Prefix:            opts.Filter,
	}
	list, err := s.s3Client.ListBuckets(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	buckets := make([]model.Bucket, len(list.Buckets))
	for i, bucket := range list.Buckets {
		buckets[i] = model.Bucket{
			Name:      bucket.Name,
			CreatedAt: bucket.CreationDate,
		}
	}
	return buckets, list.ContinuationToken, nil
}

func (s *Services) CreateBucket(ctx context.Context, name string) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	_, err := s.s3Client.CreateBucket(ctx, params)
	switch {
	case err == nil:
		return nil
	case strings.Contains(err.Error(), "BucketAlreadyOwnedByYou"):
		return errs.Conflict(errs.WithErr(ErrExistingBucket))
	default:
		return err
	}
}

func (s *Services) DeleteBucket(ctx context.Context, name string) error {
	params := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}
	_, err := s.s3Client.DeleteBucket(ctx, params)
	switch {
	case err == nil:
		return nil
	case strings.Contains(err.Error(), "BucketNotEmpty"):
		return errs.Conflict(errs.WithErr(ErrBucketNotEmpty))
	default:
		return err
	}
}

func (s *Services) PutObject(
	ctx context.Context, bucketName, objectKey, mimeType string, r io.Reader,
) (*model.Object, error) {
	params := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(mimeType),
		Body:        r,
	}
	output, err := s.s3Client.PutObject(ctx, params)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "NoSuchBucket"):
			return nil, errs.NotFound(errs.WithErr(ErrMissingBucket))
		default:
			return nil, err
		}
	}

	return &model.Object{
		Key:          &objectKey,
		Size:         output.Size,
		LastModified: aws.Time(time.Now()),
	}, nil
}

func (s *Services) DeleteObject(
	ctx context.Context, bucketName, objectKey string,
) error {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	_, err := s.s3Client.DeleteObject(ctx, params)
	return err
}

func (s *Services) GetObject(
	ctx context.Context, bucketName, objectKey string,
) (io.ReadCloser, *string, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	out, err := s.s3Client.GetObject(ctx, params)
	if err != nil {
		var opErr *smithy.OperationError
		if errors.As(err, &opErr) {
			return nil, nil, errs.NotFound(errs.WithErr(opErr.Unwrap()))
		}
		return nil, nil, err
	}
	return out.Body, out.ContentType, nil
}
