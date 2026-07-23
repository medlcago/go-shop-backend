package storage

import "time"

type UploadOptions struct {
	Metadata    map[string]string
	ContentType string
}

type TemporaryUploadURLOptions struct {
	ObjectKey   string
	ContentType string
	MinSize     int64
	MaxSize     int64
	Expires     time.Time
	Metadata    map[string]string
}

type TemporaryUploadURL struct {
	URL    string
	Fields map[string]string
}

type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
	Metadata     map[string]string
}
