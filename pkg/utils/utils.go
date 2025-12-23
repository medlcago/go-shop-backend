package utils

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/copier"
)

func Slugify(text string) string {
	return slug.Make(text)
}

func Copy(dest interface{}, src interface{}) error {
	return copier.Copy(dest, src)
}

func Ptr[T any](v T) *T {
	return &v
}
