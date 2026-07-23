package upload

import (
	"fmt"
	"slices"
	"sync"
)

type Type string

type PolicyRegistry interface {
	Get(t Type) (FilePolicy, error)
	Register(t Type, filePolicy FilePolicy)
}

type policyRegistry struct {
	policies map[Type]FilePolicy
	mu       sync.RWMutex
}

func NewPolicyRegistry() *policyRegistry {
	return &policyRegistry{
		policies: make(map[Type]FilePolicy),
	}
}
func (p *policyRegistry) Get(t Type) (FilePolicy, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	policy, ok := p.policies[t]
	if !ok {
		return FilePolicy{}, fmt.Errorf("upload policy %q not found", t)
	}

	return policy, nil
}

func (p *policyRegistry) Register(t Type, policy FilePolicy) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.policies[t]; ok {
		panic(fmt.Sprintf("upload policy %q already registered", t))
	}

	p.policies[t] = policy
}

type Format struct {
	Extensions  []string
	ContentType string
}

type FilePolicy struct {
	MinSize        int64
	MaxSize        int64
	AllowedFormats []Format
}

func (f *FilePolicy) IsValidExt(ext, ct string) bool {
	for _, f := range f.AllowedFormats {
		if f.ContentType == ct {
			return slices.Contains(f.Extensions, ext)
		}
	}
	return false
}

func (f *FilePolicy) IsValidContentType(ct string) bool {
	for _, f := range f.AllowedFormats {
		if f.ContentType == ct {
			return true
		}
	}
	return false
}
