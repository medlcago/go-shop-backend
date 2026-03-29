package upload

import (
	"fmt"
)

type Policy string

const (
	ProductImagePolicy Policy = "product_image"
)

type PolicyProvider interface {
	Register(policy Policy, fc FileConstraints) error
	Get(policy Policy) (FileConstraints, error)
}

type policyProvider struct {
	policies map[Policy]FileConstraints
}

func (p *policyProvider) Register(policy Policy, fc FileConstraints) error {
	if _, exists := p.policies[policy]; exists {
		return fmt.Errorf("policy already registered: %s", policy)
	}

	p.policies[policy] = fc
	return nil
}

func (p *policyProvider) Get(policy Policy) (FileConstraints, error) {
	c, ok := p.policies[policy]
	if !ok {
		return FileConstraints{}, fmt.Errorf("invalid policy: %s", policy)
	}
	return c, nil
}

func NewPolicyProvider() PolicyProvider {
	return &policyProvider{
		policies: make(map[Policy]FileConstraints),
	}
}
