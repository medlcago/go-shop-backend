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

func NewUploadPolicyProvider() (upload.PolicyProvider, error) {
	policyProvider := upload.NewPolicyProvider()

	if err := policyProvider.Register(upload.ProductImagePolicy, productImageConstraints); err != nil {
		return nil, fmt.Errorf("failed to register policy: %w", err)
	}

	return policyProvider, nil
}
