package upload

import (
	"errors"
	"fmt"
	"slices"
	"sync"
)

var (
	ErrPurposeNotFound          = errors.New("upload: purpose not found")
	ErrPurposeAlreadyRegistered = errors.New("upload: purpose already registered")
)

type Purpose string

const (
	ProductImagePurpose Purpose = "product-image"
)

type Registration struct {
	Policy  Policy
	Handler Handler
}

type Registry interface {
	Get(purpose Purpose) (*Registration, error)
	Register(purpose Purpose, reg *Registration)
	Purposes() []Purpose
}

type registry struct {
	regs map[Purpose]*Registration
	mu   sync.RWMutex
}

func NewRegistry() *registry {
	return &registry{regs: make(map[Purpose]*Registration)}
}

func (r *registry) Get(purpose Purpose) (*Registration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, ok := r.regs[purpose]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPurposeNotFound, purpose)
	}

	return reg, nil
}

func (r *registry) Register(purpose Purpose, reg *Registration) {
	if reg == nil {
		panic(fmt.Errorf("upload: nil registration for purpose %s", purpose))
	}

	if reg.Handler == nil {
		panic(fmt.Errorf("upload: nil handler for purpose %s", purpose))
	}

	if err := reg.Policy.validate(); err != nil {
		panic(fmt.Errorf("upload: invalid policy for purpose %s: %w", purpose, err))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.regs[purpose]; ok {
		panic(fmt.Errorf("%w: %s", ErrPurposeAlreadyRegistered, purpose))
	}

	r.regs[purpose] = reg
}

func (r *registry) Purposes() []Purpose {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Purpose, 0, len(r.regs))
	for p := range r.regs {
		out = append(out, p)
	}
	slices.Sort(out)
	return out
}
