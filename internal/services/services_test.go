package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hossein1376/s3manager/internal/model"
	"github.com/stretchr/testify/assert"
)

type mockS3Client struct {
	S3Client
	listBucketsFunc   func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	createBucketFunc  func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	listObjectsV2Func func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	deleteBucketFunc  func(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	deleteObjectsFunc func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	putObjectFunc     func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	deleteObjectFunc  func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	getObjectFunc     func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (m *mockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return m.listBucketsFunc(ctx, params, optFns...)
}

func (m *mockS3Client) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return m.createBucketFunc(ctx, params, optFns...)
}

func (m *mockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.listObjectsV2Func(ctx, params, optFns...)
}

func (m *mockS3Client) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return m.deleteBucketFunc(ctx, params, optFns...)
}

func (m *mockS3Client) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return m.deleteObjectsFunc(ctx, params, optFns...)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return m.deleteObjectFunc(ctx, params, optFns...)
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectFunc(ctx, params, optFns...)
}

func TestServices_ListObjects(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	now := time.Now()
	mockOutput := &s3.ListObjectsV2Output{
		KeyCount: aws.Int32(2),
		CommonPrefixes: []types.CommonPrefix{
			{Prefix: aws.String("dir1/")},
		},
		Contents: []types.Object{
			{Key: aws.String("file1.txt"), Size: aws.Int64(100), LastModified: &now},
		},
	}
	emptyOutput := &s3.ListObjectsV2Output{
		KeyCount: aws.Int32(0),
	}

	tests := []struct {
		name      string
		bucket    string
		maxKeys   int32
		opt       model.ListObjectsOption
		mockResp  *s3.ListObjectsV2Output
		mockErr   error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list objects success",
			bucket:    "test-bucket",
			maxKeys:   10,
			opt:       model.ListObjectsOption{},
			mockResp:  mockOutput,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "bucket not found",
			bucket:  "missing-bucket",
			maxKeys: 10,
			opt:     model.ListObjectsOption{},
			mockErr: errors.New("NoSuchBucket"),
			wantErr: true,
		},
		{
			name:      "empty bucket",
			bucket:    "empty-bucket",
			maxKeys:   10,
			opt:       model.ListObjectsOption{},
			mockResp:  emptyOutput,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "invalid bucket name",
			bucket:  "",
			maxKeys: 10,
			opt:     model.ListObjectsOption{},
			mockErr: errors.New("InvalidBucketName"),
			wantErr: true,
		},
		{
			name:      "with filter",
			bucket:    "test-bucket",
			maxKeys:   10,
			opt:       model.ListObjectsOption{Filter: "dir1/"},
			mockResp:  mockOutput,
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				listObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			s := New(mock)
			got, _, err := s.ListObjects(context.Background(), tt.bucket, tt.maxKeys, tt.opt)
			a.Equal(tt.wantErr, err != nil)
			if err == nil {
				a.Len(got, tt.wantCount)
			}
		})
	}
}

func TestServices_ListBuckets(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	now := time.Now()
	mockBuckets := []types.Bucket{
		{Name: aws.String("bucket-1"), CreationDate: &now},
		{Name: aws.String("bucket-2"), CreationDate: &now},
		{Name: aws.String("other-bucket"), CreationDate: &now},
	}
	emptyBuckets := []types.Bucket{}

	tests := []struct {
		name      string
		count     int32
		opts      model.ListBucketsOptions
		mockResp  *s3.ListBucketsOutput
		mockErr   error
		wantCount int
		wantNext  bool
		wantErr   bool
	}{
		{
			name:  "list all buckets",
			count: 10,
			opts:  model.ListBucketsOptions{},
			mockResp: &s3.ListBucketsOutput{
				Buckets: mockBuckets,
			},
			wantCount: 3,
			wantNext:  false,
			wantErr:   false,
		},
		{
			name:  "filter buckets by prefix",
			count: 10,
			opts:  model.ListBucketsOptions{Filter: aws.String("bucket")},
			mockResp: &s3.ListBucketsOutput{
				Buckets: mockBuckets,
			},
			wantCount: 2,
			wantNext:  false,
			wantErr:   false,
		},
		{
			name:  "pagination",
			count: 1,
			opts:  model.ListBucketsOptions{},
			mockResp: &s3.ListBucketsOutput{
				Buckets: mockBuckets,
			},
			wantCount: 1,
			wantNext:  true,
			wantErr:   false,
		},
		{
			name:  "no buckets",
			count: 10,
			opts:  model.ListBucketsOptions{},
			mockResp: &s3.ListBucketsOutput{
				Buckets: emptyBuckets,
			},
			wantCount: 0,
			wantNext:  false,
			wantErr:   false,
		},
		{
			name:    "access denied",
			count:   10,
			opts:    model.ListBucketsOptions{},
			mockErr: errors.New("AccessDenied"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				listBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			s := New(mock)
			got, next, err := s.ListBuckets(context.Background(), tt.count, tt.opts)
			a.Equal(tt.wantErr, err != nil)
			if err == nil {
				a.Len(got, tt.wantCount)
				a.Equal(tt.wantNext, next != nil)
			}
		})
	}
}

func TestServices_CreateBucket(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tests := []struct {
		name    string
		bucket  string
		mockErr error
		wantErr error
	}{
		{
			name:    "success",
			bucket:  "new-bucket",
			mockErr: nil,
			wantErr: nil,
		},
		{
			name:    "already exists",
			bucket:  "existing-bucket",
			mockErr: errors.New("BucketAlreadyOwnedByYou"),
			wantErr: errors.New("Conflict"),
		},
		{
			name:    "invalid name",
			bucket:  "invalid_name",
			mockErr: errors.New("InvalidBucketName"),
			wantErr: errors.New("Bad Request"),
		},
		{
			name:    "empty name",
			bucket:  "",
			mockErr: errors.New("InvalidBucketName"),
			wantErr: errors.New("Bad Request"),
		},
		{
			name:    "access denied",
			bucket:  "denied-bucket",
			mockErr: errors.New("AccessDenied"),
			wantErr: errors.New("Forbidden"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				createBucketFunc: func(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
					return &s3.CreateBucketOutput{}, tt.mockErr
				},
			}
			s := New(mock)
			err := s.CreateBucket(context.Background(), tt.bucket)
			if tt.wantErr != nil {
				a.Error(err)
				a.Contains(err.Error(), tt.wantErr.Error())
			} else {
				a.NoError(err)
			}
		})
	}
}

func TestServices_DeleteBucket(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tests := []struct {
		name      string
		bucket    string
		recursive bool
		mockErr   error
		wantErr   error
	}{
		{
			name:      "success",
			bucket:    "test-bucket",
			recursive: false,
			mockErr:   nil,
			wantErr:   nil,
		},
		{
			name:      "not empty",
			bucket:    "full-bucket",
			recursive: false,
			mockErr:   errors.New("BucketNotEmpty"),
			wantErr:   errors.New("Conflict"),
		},
		{
			name:      "not found",
			bucket:    "missing-bucket",
			recursive: false,
			mockErr:   errors.New("NoSuchBucket"),
			wantErr:   errors.New("Not Found"),
		},
		{
			name:      "empty name",
			bucket:    "",
			recursive: false,
			mockErr:   errors.New("InvalidBucketName"),
			wantErr:   errors.New("Bad Request"),
		},
		{
			name:      "recursive delete",
			bucket:    "full-bucket",
			recursive: true,
			mockErr:   nil,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				deleteBucketFunc: func(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
					return &s3.DeleteBucketOutput{}, tt.mockErr
				},
			}
			s := New(mock)
			err := s.DeleteBucket(context.Background(), tt.bucket, tt.recursive)
			if tt.wantErr != nil {
				a.Error(err)
				a.Contains(err.Error(), tt.wantErr.Error())
			} else {
				a.NoError(err)
			}
		})
	}
}

func TestServices_PutObject(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tests := []struct {
		name        string
		bucket      string
		key         string
		contentType string
		content     string
		mockErr     error
		wantErr     bool
	}{
		{
			name:        "success",
			bucket:      "test-bucket",
			key:         "test.txt",
			contentType: "text/plain",
			content:     "hello world",
			wantErr:     false,
		},
		{
			name:        "empty content",
			bucket:      "test-bucket",
			key:         "empty.txt",
			contentType: "text/plain",
			content:     "",
			wantErr:     false,
		},
		{
			name:        "large content",
			bucket:      "test-bucket",
			key:         "large.txt",
			contentType: "application/octet-stream",
			content:     strings.Repeat("a", 1000), // 1KB
			wantErr:     false,
		},
		{
			name:        "bucket not found",
			bucket:      "missing-bucket",
			key:         "test.txt",
			contentType: "text/plain",
			content:     "hello",
			mockErr:     errors.New("NoSuchBucket"),
			wantErr:     true,
		},
		{
			name:        "access denied",
			bucket:      "denied-bucket",
			key:         "test.txt",
			contentType: "text/plain",
			content:     "hello",
			mockErr:     errors.New("AccessDenied"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					return &s3.PutObjectOutput{Size: aws.Int64(int64(len(tt.content)))}, tt.mockErr
				},
			}
			s := New(mock)
			got, err := s.PutObject(context.Background(), tt.bucket, tt.key, tt.contentType, strings.NewReader(tt.content))
			a.Equal(tt.wantErr, err != nil)
			if err == nil {
				a.Equal(tt.key, *got.Key)
			}
		})
	}
}

func TestServices_DeleteObject(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tests := []struct {
		name      string
		bucket    string
		key       string
		recursive bool
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "delete object success",
			bucket:    "test-bucket",
			key:       "file.txt",
			recursive: false,
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "object not found",
			bucket:    "test-bucket",
			key:       "missing.txt",
			recursive: false,
			mockErr:   nil, // S3 delete is idempotent
			wantErr:   false,
		},
		{
			name:      "bucket not found",
			bucket:    "missing-bucket",
			key:       "file.txt",
			recursive: false,
			mockErr:   errors.New("NoSuchBucket"),
			wantErr:   true,
		},
		{
			name:      "empty key",
			bucket:    "test-bucket",
			key:       "",
			recursive: false,
			mockErr:   errors.New("KeyTooLongError"), // or appropriate error
			wantErr:   true,
		},
		{
			name:      "recursive delete",
			bucket:    "test-bucket",
			key:       "dir/",
			recursive: true,
			mockErr:   nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				listObjectsV2Func: func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					if tt.recursive {
						return &s3.ListObjectsV2Output{
							Contents: []types.Object{
								{Key: aws.String("dir/file1.txt")},
								{Key: aws.String("dir/file2.txt")},
							},
						}, nil
					}
					return &s3.ListObjectsV2Output{}, nil
				},
				deleteObjectFunc: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
					return &s3.DeleteObjectOutput{}, tt.mockErr
				},
				deleteObjectsFunc: func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
					return &s3.DeleteObjectsOutput{}, tt.mockErr
				},
			}
			s := New(mock)
			err := s.DeleteObject(context.Background(), tt.bucket, tt.key, tt.recursive)
			a.Equal(tt.wantErr, err != nil)
		})
	}
}

func BenchmarkServices_ListBuckets(b *testing.B) {
	now := time.Now()
	mockBuckets := []types.Bucket{
		{Name: aws.String("bucket-1"), CreationDate: &now},
		{Name: aws.String("bucket-2"), CreationDate: &now},
		{Name: aws.String("bucket-3"), CreationDate: &now},
	}

	mock := &mockS3Client{
		listBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			return &s3.ListBucketsOutput{
				Buckets: mockBuckets,
			}, nil
		},
	}
	s := New(mock)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = s.ListBuckets(context.Background(), 10, model.ListBucketsOptions{})
	}
}

func TestServices_GetObject(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tests := []struct {
		name         string
		bucket       string
		key          string
		mockContent  string
		mockErr      error
		wantErr      bool
		expectedType string
	}{
		{
			name:         "success",
			bucket:       "test-bucket",
			key:          "test.txt",
			mockContent:  "hello world",
			expectedType: "text/plain",
			wantErr:      false,
		},
		{
			name:        "empty object",
			bucket:      "test-bucket",
			key:         "empty.txt",
			mockContent: "",
			wantErr:     false,
		},
		{
			name:    "object not found",
			bucket:  "test-bucket",
			key:     "missing.txt",
			mockErr: errors.New("NoSuchKey"),
			wantErr: true,
		},
		{
			name:    "bucket not found",
			bucket:  "missing-bucket",
			key:     "test.txt",
			mockErr: errors.New("NoSuchBucket"),
			wantErr: true,
		},
		{
			name:    "access denied",
			bucket:  "denied-bucket",
			key:     "test.txt",
			mockErr: errors.New("AccessDenied"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{
				getObjectFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return &s3.GetObjectOutput{
						Body:        io.NopCloser(strings.NewReader(tt.mockContent)),
						ContentType: aws.String(tt.expectedType),
					}, nil
				},
			}
			s := New(mock)
			body, contentType, err := s.GetObject(context.Background(), tt.bucket, tt.key)
			a.Equal(tt.wantErr, err != nil)
			if err == nil {
				defer body.Close()
				content, err := io.ReadAll(body)
				a.NoError(err)
				a.Equal(tt.mockContent, string(content))
				if tt.expectedType != "" {
					a.Equal(tt.expectedType, *contentType)
				}
			}
		})
	}
}
