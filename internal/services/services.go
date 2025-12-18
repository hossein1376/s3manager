package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/s3manager/internal/model"
)

var (
	ErrBucketNotEmpty = errors.New("bucket is not empty")
	ErrExistingBucket = errors.New("bucket already exists")
	ErrMissingBucket  = errors.New("bucket not found")
	ErrDirNotEmpty    = errors.New("directory is not empty")
	ErrInvalidName    = errors.New("invalid bucket name")
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
	prefix := opt.Path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	pathPrefix := prefix
	prefix += opt.Filter
	params := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucketName),
		MaxKeys:           aws.Int32(maxKeys),
		ContinuationToken: opt.ContinuationToken,
		Prefix:            &prefix,
		Delimiter:         aws.String("/"),
	}
	list, err := s.s3Client.ListObjectsV2(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchBucket") {
			return nil, nil, errs.NotFound(errs.WithMsg(ErrMissingBucket.Error()))
		}
		return nil, nil, err
	}
	keyCount := int32(0)
	if list.KeyCount != nil {
		keyCount = *list.KeyCount
	}
	objects := make([]model.Object, 0, keyCount)
	for _, obj := range list.CommonPrefixes {
		var key *string
		if obj.Prefix != nil {
			key = aws.String(
				strings.TrimSuffix(strings.TrimPrefix(*obj.Prefix, pathPrefix), "/"),
			)
		}
		objects = append(objects, model.Object{
			Key:   key,
			IsDir: true,
		})
	}
	for _, obj := range list.Contents {
		var lastModified *string
		var key *string
		if obj.LastModified != nil {
			lastModified = aws.String(obj.LastModified.Format(time.DateTime))
		}
		if obj.Key != nil {
			key = aws.String(strings.TrimPrefix(*obj.Key, pathPrefix))
		}
		objects = append(objects, model.Object{
			Key:          key,
			Size:         obj.Size,
			LastModified: lastModified,
		})
	}
	return objects, list.NextContinuationToken, nil
}

func (s *Services) ListBuckets(
	ctx context.Context, count int32, opts model.ListBucketsOptions,
) ([]model.Bucket, *string, error) {
	// List all buckets since S3 doesn't support prefix filtering
	var allBuckets []model.Bucket
	var continuationToken *string
	for {
		params := &s3.ListBucketsInput{
			ContinuationToken: continuationToken,
		}
		list, err := s.s3Client.ListBuckets(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		for _, bucket := range list.Buckets {
			var createdAt *string
			if bucket.CreationDate != nil {
				createdAt = aws.String(bucket.CreationDate.Format(time.DateTime))
			}
			allBuckets = append(allBuckets, model.Bucket{
				Name:      bucket.Name,
				CreatedAt: createdAt,
			})
		}
		if list.ContinuationToken == nil {
			break
		}
		continuationToken = list.ContinuationToken
	}

	// Filter by prefix if provided
	if opts.Filter != nil && *opts.Filter != "" {
		filtered := make([]model.Bucket, 0)
		for _, b := range allBuckets {
			if b.Name != nil && strings.HasPrefix(*b.Name, *opts.Filter) {
				filtered = append(filtered, b)
			}
		}
		allBuckets = filtered
	}

	// Apply pagination
	start := 0
	if opts.ContinuationToken != nil {
		// For simplicity, assume token is index as string
		if idx, err := strconv.Atoi(*opts.ContinuationToken); err == nil {
			start = idx
		}
	}
	if start < 0 {
		start = 0
	}
	if start > len(allBuckets) {
		start = len(allBuckets)
	}
	end := start + int(count)
	if end > len(allBuckets) {
		end = len(allBuckets)
	}
	result := allBuckets[start:end]

	var nextToken *string
	if end < len(allBuckets) {
		nextToken = aws.String(strconv.Itoa(end))
	}

	return result, nextToken, nil
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
		return errs.Conflict(errs.WithMsg(ErrExistingBucket.Error()))
	case strings.Contains(err.Error(), "InvalidBucketName"):
		return errs.BadRequest(errs.WithMsg(ErrInvalidName.Error()))
	default:
		return err
	}
}

func (s *Services) DeleteBucket(ctx context.Context, name string, recursive bool) error {
	if recursive {
		// Delete all objects in the bucket first
		err := s.deleteAllObjects(ctx, name)
		if err != nil {
			return fmt.Errorf("deleting objects in bucket: %w", err)
		}
	}
	params := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}
	_, err := s.s3Client.DeleteBucket(ctx, params)
	switch {
	case err == nil:
		return nil
	case strings.Contains(err.Error(), "BucketNotEmpty"):
		return errs.Conflict(errs.WithMsg(ErrBucketNotEmpty.Error()))
	default:
		return fmt.Errorf("deleting bucket: %w", err)
	}
}

func (s *Services) deleteAllObjects(ctx context.Context, bucketName string) error {
	// List all objects
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	for {
		list, err := s.s3Client.ListObjectsV2(ctx, params)
		if err != nil {
			return fmt.Errorf("listing objects for deletion: %w", err)
		}
		// Delete objects in batches
		if len(list.Contents) > 0 {
			var objectsToDelete []types.ObjectIdentifier
			for _, obj := range list.Contents {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
			}

			_, err = s.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &types.Delete{
					Objects: objectsToDelete,
					Quiet:   aws.Bool(true),
				},
			})
			if err != nil {
				return fmt.Errorf("deleting objects: %w", err)
			}
		}
		if list.NextContinuationToken == nil {
			break
		}
		params.ContinuationToken = list.NextContinuationToken
	}
	return nil
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
			return nil, errs.NotFound(errs.WithMsg(ErrMissingBucket.Error()))
		default:
			return nil, err
		}
	}

	return &model.Object{
		Key:          &objectKey,
		Size:         output.Size,
		LastModified: aws.String(time.Now().Format(time.DateTime)),
	}, nil
}

func (s *Services) DeleteObject(
	ctx context.Context, bucketName, objectKey string, recursive bool,
) error {
	if recursive {
		return s.deleteObjectsWithPrefix(ctx, bucketName, objectKey)
	}

	// Check if attempting to delete a non-empty directory without recursive flag
	prefix := objectKey
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	listParams := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1), // Just check if any objects exist
	}
	list, err := s.s3Client.ListObjectsV2(ctx, listParams)
	switch {
	case err != nil:
		return fmt.Errorf("checking for objects: %w", err)
	case len(list.Contents) > 0:
		return errs.BadRequest(errs.WithMsg(ErrDirNotEmpty.Error()))
	}

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	_, err = s.s3Client.DeleteObject(ctx, params)
	return err
}

func (s *Services) deleteObjectsWithPrefix(
	ctx context.Context, bucketName, prefix string,
) error {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}
	for {
		list, err := s.s3Client.ListObjectsV2(ctx, params)
		if err != nil {
			return fmt.Errorf("listing objects for deletion: %w", err)
		}
		// Delete objects in batches
		if len(list.Contents) > 0 {
			var objectsToDelete []types.ObjectIdentifier
			for _, obj := range list.Contents {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
			}

			_, err = s.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &types.Delete{
					Objects: objectsToDelete,
					Quiet:   aws.Bool(true),
				},
			})
			if err != nil {
				return fmt.Errorf("deleting objects: %w", err)
			}
		}
		if list.NextContinuationToken == nil {
			break
		}
		params.ContinuationToken = list.NextContinuationToken
	}
	return nil
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
			return nil, nil, errs.NotFound(
				errs.WithErr(opErr.Unwrap()), errs.WithMsg("object not found"),
			)
		}
		return nil, nil, err
	}
	return out.Body, out.ContentType, nil
}
