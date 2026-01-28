package contenttype

import (
	"fmt"
	"io"
	"net/http"
)

type MagicDetector struct{}

func (m *MagicDetector) Detect(r io.ReadSeeker) (contentType string, err error) {
	const op = "MagicDetector.Detect"

	defer func() {
		_, err = r.Seek(0, io.SeekStart)
		if err != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}()

	buff := make([]byte, 512)
	n, err := r.Read(buff)
	if err != nil && err != io.EOF {
		err = fmt.Errorf("%s: %w", op, err)
		return
	}

	contentType = http.DetectContentType(buff[:n])
	return
}

func NewMagicDetector() *MagicDetector {
	return &MagicDetector{}
}
