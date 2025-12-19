package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/s3manager/internal/config"
	"github.com/hossein1376/s3manager/internal/model"
	"github.com/stretchr/testify/assert"
)

// setupHandler creates a test handler with a mock service
func setupHandler(svc *mockService) *Handler {
	return &Handler{
		cfg:     config.Config{},
		service: svc,
	}
}

type mockService struct {
	listBucketsFunc  func(ctx context.Context, count int32, opts model.ListBucketsOptions) ([]model.Bucket, *string, error)
	listObjectsFunc  func(ctx context.Context, bucketName string, maxKeys int32, opt model.ListObjectsOption) ([]model.Object, *string, error)
	createBucketFunc func(ctx context.Context, name string) error
	deleteBucketFunc func(ctx context.Context, name string, recursive bool) error
	putObjectFunc    func(ctx context.Context, bucketName, objectKey, mimeType string, r io.Reader) (*model.Object, error)
	deleteObjectFunc func(ctx context.Context, bucketName, objectKey string, recursive bool) error
	getObjectFunc    func(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, *string, error)
}

func (m *mockService) ListBuckets(ctx context.Context, count int32, opts model.ListBucketsOptions) ([]model.Bucket, *string, error) {
	return m.listBucketsFunc(ctx, count, opts)
}

func (m *mockService) ListObjects(ctx context.Context, bucketName string, maxKeys int32, opt model.ListObjectsOption) ([]model.Object, *string, error) {
	return m.listObjectsFunc(ctx, bucketName, maxKeys, opt)
}

func (m *mockService) CreateBucket(ctx context.Context, name string) error {
	return m.createBucketFunc(ctx, name)
}

func (m *mockService) DeleteBucket(ctx context.Context, name string, recursive bool) error {
	return m.deleteBucketFunc(ctx, name, recursive)
}

func (m *mockService) PutObject(ctx context.Context, bucketName, objectKey, mimeType string, r io.Reader) (*model.Object, error) {
	return m.putObjectFunc(ctx, bucketName, objectKey, mimeType, r)
}

func (m *mockService) DeleteObject(ctx context.Context, bucketName, objectKey string, recursive bool) error {
	return m.deleteObjectFunc(ctx, bucketName, objectKey, recursive)
}

func (m *mockService) GetObject(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, *string, error) {
	return m.getObjectFunc(ctx, bucketName, objectKey)
}

func TestHandler_ListBucketsHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	mockBuckets := []model.Bucket{
		{Name: aws.String("bucket-1"), CreatedAt: aws.String("2023-01-01 10:00:00")},
		{Name: aws.String("bucket-2"), CreatedAt: aws.String("2023-01-02 11:00:00")},
	}

	svc := &mockService{
		listBucketsFunc: func(ctx context.Context, count int32, opts model.ListBucketsOptions) ([]model.Bucket, *string, error) {
			return mockBuckets, nil, nil
		},
	}

	h := setupHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/buckets", nil)
	w := httptest.NewRecorder()

	h.ListBucketsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	a.Equal(http.StatusOK, res.StatusCode)

	var response struct {
		Buckets []model.Bucket `json:"buckets"`
	}
	err := json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Len(response.Buckets, len(mockBuckets))
	a.Equal(*mockBuckets[0].Name, *response.Buckets[0].Name)
}

func TestHandler_CreateBucketHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		createBucketFunc: func(ctx context.Context, name string) error {
			if name == "new-bucket" {
				return nil
			}
			return errors.New("error")
		},
	}

	h := setupHandler(svc)

	body := `{"name": "new-bucket"}`
	req := httptest.NewRequest(http.MethodPost, "/api/buckets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateBucketHandler(w, req)

	res := w.Result()
	a.Equal(http.StatusNoContent, res.StatusCode)
}

func TestHandler_ListObjectsHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		listObjectsFunc: func(ctx context.Context, bucketName string, maxKeys int32, opt model.ListObjectsOption) ([]model.Object, *string, error) {
			if bucketName != "test-bucket" {
				return nil, nil, errors.New("wrong bucket")
			}
			return []model.Object{{Key: aws.String("file.txt")}}, nil, nil
		},
	}

	h := setupHandler(svc)
	r := newRouter(h, nil, true)

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/test-bucket", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	a.Equal(http.StatusOK, res.StatusCode)

	var response struct {
		List []model.Object `json:"list"`
	}
	err := json.NewDecoder(res.Body).Decode(&response)
	a.NoError(err)

	a.Len(response.List, 1)
}

func TestHandler_DeleteBucketHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		deleteBucketFunc: func(ctx context.Context, name string, recursive bool) error {
			if name == "test-bucket" {
				return nil
			}
			return errors.New("not found")
		},
	}

	h := setupHandler(svc)
	r := newRouter(h, nil, true)

	req := httptest.NewRequest(http.MethodDelete, "/api/buckets/test-bucket", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	a.Equal(http.StatusNoContent, res.StatusCode)
}

func TestHandler_PutObjectHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		putObjectFunc: func(ctx context.Context, bucketName, objectKey, mimeType string, r io.Reader) (*model.Object, error) {
			return &model.Object{Key: &objectKey}, nil
		},
	}

	h := &Handler{
		cfg:     config.Config{S3: config.S3{MaxSizeBytes: 1024 * 1024}},
		service: svc,
	}
	r := newRouter(h, nil, true)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	a.NoError(err)
	_, err = part.Write([]byte("content"))
	a.NoError(err)
	err = writer.WriteField("key", "test.txt")
	a.NoError(err)
	err = writer.Close()
	a.NoError(err)

	req := httptest.NewRequest(http.MethodPut, "/api/buckets/test-bucket/objects", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	a.Equal(http.StatusCreated, res.StatusCode)
}

func TestHandler_GetObjectHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	mockContent := "hello"
	svc := &mockService{
		getObjectFunc: func(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, *string, error) {
			return io.NopCloser(strings.NewReader(mockContent)), aws.String("text/plain"), nil
		},
	}

	h := setupHandler(svc)
	r := newRouter(h, nil, true)

	req := httptest.NewRequest(http.MethodGet, "/api/buckets/test-bucket/objects/test.txt", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	a.Equal(http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	a.NoError(err)
	a.Equal(mockContent, string(body))
}

func TestHandler_DeleteObjectHandler(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		deleteObjectFunc: func(ctx context.Context, bucketName, objectKey string, recursive bool) error {
			return nil
		},
	}

	h := setupHandler(svc)
	r := newRouter(h, nil, true)

	req := httptest.NewRequest(http.MethodDelete, "/api/buckets/test-bucket/objects/test.txt", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	a.Equal(http.StatusNoContent, res.StatusCode)
}

func TestHandler_ListBucketsHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		listBucketsFunc: func(ctx context.Context, count int32, opts model.ListBucketsOptions) ([]model.Bucket, *string, error) {
			return nil, nil, errors.New("service error")
		},
	}

	h := setupHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/buckets", nil)
	w := httptest.NewRecorder()

	h.ListBucketsHandler(w, req)

	res := w.Result()
	a.Equal(http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_CreateBucketHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		createBucketFunc: func(ctx context.Context, name string) error {
			return errors.New("invalid name")
		},
	}

	h := setupHandler(svc)
	body := `{"name": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/buckets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateBucketHandler(w, req)

	res := w.Result()
	a.Equal(http.StatusBadRequest, res.StatusCode)
}

func TestHandler_PutObjectHandler_TooLarge(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	svc := &mockService{
		putObjectFunc: func(ctx context.Context, bucketName, objectKey, mimeType string, r io.Reader) (*model.Object, error) {
			return nil, errs.New(http.StatusRequestEntityTooLarge, errs.WithMsg("file too large"))
		},
	}

	h := &Handler{
		cfg:     config.Config{S3: config.S3{MaxSizeBytes: 100}},
		service: svc,
	}
	r := newRouter(h, nil, true)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "large.txt")
	a.NoError(err)
	_, err = part.Write([]byte(strings.Repeat("a", 200)))
	a.NoError(err)
	err = writer.WriteField("key", "large.txt")
	a.NoError(err)
	err = writer.Close()
	a.NoError(err)

	req := httptest.NewRequest(http.MethodPut, "/api/buckets/test-bucket/objects", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	a.Equal(http.StatusRequestEntityTooLarge, res.StatusCode)
}
