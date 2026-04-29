package contenttype

import (
	"fmt"
	"io"
	"net/http"
)

type MagicDetector struct{}

func (m *MagicDetector) Detect(r io.ReadSeeker) (contentType string, err error) {
	defer func() {
		_, err = r.Seek(0, io.SeekStart)
		if err != nil {
			err = fmt.Errorf("magic content type detect: %w", err)
		}
	}()

	buff := make([]byte, 512)
	n, err := r.Read(buff)
	if err != nil && err != io.EOF {
		err = fmt.Errorf("magic content type detect: %w", err)
		return
	}

	contentType = http.DetectContentType(buff[:n])
	return
}

func NewMagicDetector() *MagicDetector {
	return &MagicDetector{}
}
