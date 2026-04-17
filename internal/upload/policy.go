package upload

import (
	"fmt"
)

type Policy string

const (
	ProductImagePolicy Policy = "product_image"
)

type PolicyProvider interface {
	Get(policy Policy) (FileConstraints, error)
}

type policyProvider struct {
	policies map[Policy]FileConstraints
}

func (p *policyProvider) Get(policy Policy) (FileConstraints, error) {
	c, ok := p.policies[policy]
	if !ok {
		return FileConstraints{}, fmt.Errorf("unknown policy: %s", policy)
	}
	return c, nil
}

type PolicyEntry struct {
	Policy      Policy
	Constraints FileConstraints
}

func NewPolicyProvider(entries ...PolicyEntry) (PolicyProvider, error) {
	p := &policyProvider{policies: make(map[Policy]FileConstraints, len(entries))}
	for _, e := range entries {
		if _, exists := p.policies[e.Policy]; exists {
			return nil, fmt.Errorf("duplicate policy: %s", e.Policy)
		}
		p.policies[e.Policy] = e.Constraints
	}

	return p, nil
}
