package utils

import (
	"io"
	"mime"
	"net/http"

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

func DetectContentType(r io.ReadSeeker) (string, error) {
	buff := make([]byte, 512)
	_, err := r.Read(buff)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buff)

	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	return contentType, nil
}

func ContentTypeToExt(contentType string) ([]string, error) {
	baseType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	extensions, err := mime.ExtensionsByType(baseType)
	if err != nil {
		return nil, err
	}

	return extensions, nil
}
