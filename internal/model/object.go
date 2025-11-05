package model

type Object struct {
	Key          *string `json:"key"`
	IsDir        bool    `json:"is_dir"`
	Size         *int64  `json:"size,omitempty"`
	LastModified *string `json:"last_modified,omitempty"`
}

type ListObjectsOption struct {
	Path              string
	Filter            string
	ContinuationToken *string
}
