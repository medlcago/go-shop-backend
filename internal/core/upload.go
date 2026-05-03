package core

import (
	"fmt"
	"go-shop-backend/internal/upload"
)

var (
	productImageConstraints = upload.FileConstraints{
		MaxSize: 5 << 20,
		AllowedFormats: []upload.Format{
			{
				Extensions: []string{"jpg", "jpeg"}, ContentType: "image/jpeg",
			},
			{
				Extensions: []string{"png"}, ContentType: "image/png",
			},
		},
	}
)

var (
	productImagePolicyEntry = upload.PolicyEntry{
		Policy:      upload.ProductImagePolicy,
		Constraints: productImageConstraints,
	}
)

func NewUploadPolicyProvider() (upload.PolicyProvider, error) {
	policyEntries := []upload.PolicyEntry{
		productImagePolicyEntry,
	}

	policyProvider, err := upload.NewPolicyProvider(policyEntries...)
	if err != nil {
		return nil, fmt.Errorf("core: upload.NewPolicyProvider failed: %w", err)
	}

	return policyProvider, nil
}
