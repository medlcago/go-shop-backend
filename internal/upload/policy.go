package upload

import (
	"errors"
	"fmt"
	"slices"
	"sync"
)

var (
	ErrPolicyNotFound        = errors.New("upload policy not found")
	ErrTypeAlreadyRegistered = errors.New("upload type already registered")
)

type Type string

type Registry interface {
	Get(t Type) (Policy, error)
	Register(t Type, policy Policy)
}

type registry struct {
	policies map[Type]Policy
	mu       sync.RWMutex
}

func NewRegistry() *registry {
	r := &registry{
		policies: make(map[Type]Policy),
	}

	return r
}

func (r *registry) Get(t Type) (Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policy, ok := r.policies[t]
	if !ok {
		return Policy{}, fmt.Errorf("%w: %s", ErrPolicyNotFound, t)
	}

	return policy, nil
}

func (r *registry) Register(t Type, policy Policy) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.policies[t]; ok {
		panic(fmt.Errorf("%w: %s", ErrTypeAlreadyRegistered, t))
	}

	r.policies[t] = policy
}

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
