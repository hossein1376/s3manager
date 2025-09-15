package model

import "time"

type Object struct {
	Key          *string    `json:"key"`
	Size         *int64     `json:"size"`
	LastModified *time.Time `json:"last_modified"`
}

type ListObjectsOption struct {
	Filter            *string
	ContinuationToken *string
}
