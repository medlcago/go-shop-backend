package upload

import (
	"slices"
)

type Format struct {
	Extensions  []string
	ContentType string
}

type FileConstraints struct {
	MaxSize        int64
	AllowedFormats []Format
}

func (c *FileConstraints) IsValidExt(ext, ct string) bool {
	for _, f := range c.AllowedFormats {
		if f.ContentType == ct {
			return slices.Contains(f.Extensions, ext)
		}
	}
	return false
}

func (c *FileConstraints) IsValidType(ct string) bool {
	for _, f := range c.AllowedFormats {
		if f.ContentType == ct {
			return true
		}
	}
	return false
}
