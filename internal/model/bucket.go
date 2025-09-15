package model

import "time"

type Bucket struct {
	Name      *string    `json:"name"`
	CreatedAt *time.Time `json:"created_at"`
}

type ListBucketsOptions struct {
	Filter            *string
	ContinuationToken *string
}
