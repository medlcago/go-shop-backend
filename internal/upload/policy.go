package upload

import (
	"errors"
	"fmt"
	"slices"
)

type Format struct {
	Extensions  []string
	ContentType string
}

type Policy struct {
	MinSize        int64
	MaxSize        int64
	AllowedFormats []Format
}

func (p *Policy) CalculateEffectiveMaxSize(maxSize int64) int64 {
	return min(p.MaxSize, maxSize)
}

func (p *Policy) IsValidExt(ext, ct string) bool {
	for _, f := range p.AllowedFormats {
		if f.ContentType == ct {
			return slices.Contains(f.Extensions, ext)
		}
	}
	return false
}

func (p *Policy) IsValidContentType(ct string) bool {
	for _, f := range p.AllowedFormats {
		if f.ContentType == ct {
			return true
		}
	}
	return false
}

func (p *Policy) validate() error {
	if p.MinSize < 0 {
		return fmt.Errorf("min size must be >= 0, got %d", p.MinSize)
	}

	if p.MaxSize <= 0 {
		return fmt.Errorf("max size must be > 0, got %d", p.MaxSize)
	}

	if p.MinSize > p.MaxSize {
		return fmt.Errorf("min size %d > max size %d", p.MinSize, p.MaxSize)
	}

	if len(p.AllowedFormats) == 0 {
		return errors.New("allowed formats must not be empty")
	}

	for i, f := range p.AllowedFormats {
		if f.ContentType == "" {
			return fmt.Errorf("format[%d]: empty content type", i)
		}

		if len(f.Extensions) == 0 {
			return fmt.Errorf("format[%d]: no extensions", i)
		}
	}

	return nil
}
