package upload

import (
	"errors"
	"fmt"
	"go-shop-backend/pkg/apperror"
	"slices"
)

type File struct {
	Size        int64
	ContentType string
}

type Policy struct {
	MinSize      int64
	MaxSize      int64
	MaxSizeLimit int64
	AllowedMimes []string
}

func NewPolicy(
	minSize int64,
	maxSize int64,
	maxSizeLimit int64,
	mimes []string,
) (Policy, error) {
	effectiveMaxSize := calculateEffectiveMaxSize(maxSize, maxSizeLimit)

	p := Policy{
		MinSize:      minSize,
		MaxSize:      effectiveMaxSize,
		AllowedMimes: slices.Clone(mimes),
	}

	if err := p.validate(); err != nil {
		return Policy{}, err
	}

	return p, nil
}

func (p *Policy) Validate(f File) error {
	if f.Size < p.MinSize {
		return apperror.ErrInvalidFileSize
	}

	if f.Size > p.MaxSize {
		return apperror.ErrInvalidFileSize
	}

	if !slices.Contains(p.AllowedMimes, f.ContentType) {
		return apperror.ErrInvalidFileType
	}

	return nil
}

func calculateEffectiveMaxSize(maxSize int64, maxSizeLimit int64) int64 {
	effectiveMax := maxSize
	if maxSizeLimit > 0 && (effectiveMax <= 0 || maxSizeLimit < effectiveMax) {
		effectiveMax = maxSizeLimit
	}
	return effectiveMax
}

func (p *Policy) validate() error {
	if p.MinSize <= 0 {
		return fmt.Errorf("policy: min size must be > 0, got %d", p.MinSize)
	}

	if p.MaxSize <= 0 {
		return fmt.Errorf("policy: max size must be > 0, got %d", p.MaxSize)
	}

	if p.MinSize > p.MaxSize {
		return fmt.Errorf("policy: min size %d > max size %d", p.MinSize, p.MaxSize)
	}

	if len(p.AllowedMimes) == 0 {
		return errors.New("policy: allowed mimes must not be empty")
	}

	return nil
}
