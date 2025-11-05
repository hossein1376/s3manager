package model

type Bucket struct {
	Name      *string `json:"name"`
	CreatedAt *string `json:"created_at"`
}

type ListBucketsOptions struct {
	Filter            *string
	ContinuationToken *string
}
