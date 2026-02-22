package contenttype

import "io"

type Detector interface {
	Detect(r io.ReadSeeker) (string, error)
}
