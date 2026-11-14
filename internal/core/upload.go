package core

import (
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/internal/upload"
	uploadHandlers "go-shop-backend/internal/upload/handlers"
	"reflect"
)

func newUploadHandlers(container *Container) map[upload.Purpose]upload.Handler {
	handlers := make(map[upload.Purpose]upload.Handler)

	productImageHandler := uploadHandlers.NewProductImageHandler(
		container.ProductRepo(),
		container.UploadRepo(),
		container.Config().Upload.MaxImagesPerProduct,
	)

	handlers[upload.ProductImagePurpose] = productImageHandler

	return handlers
}

func newUploadPolicies(uploadConfig config.Upload) (map[upload.Purpose]upload.Policy, error) {
	policies := make(map[upload.Purpose]upload.Policy)

	productImagePolicy, err := upload.NewPolicy(
		5<<10, // 5 KB
		5<<20, // 5 MB,
		uploadConfig.MaxFileSize,
		[]string{"image/jpeg", "image/png"},
	)
	if err != nil {
		return nil, err
	}

	policies[upload.ProductImagePurpose] = productImagePolicy

	return policies, nil
}

func newUploadRegistry(container *Container) (upload.Registry, error) {
	policies, err := newUploadPolicies(container.Config().Upload)
	if err != nil {
		return nil, err
	}

	handlers := newUploadHandlers(container)

	if len(handlers) != len(policies) {
		return nil, fmt.Errorf(
			"handlers/policies mismatch: %d handlers, %d policies",
			len(handlers), len(policies),
		)
	}

	registry := upload.NewRegistry()

	for purpose, handler := range handlers {
		policy, ok := policies[purpose]
		if !ok || reflect.ValueOf(policy).IsZero() {
			return nil, fmt.Errorf("no policy for purpose %q", purpose)
		}

		registry.Register(
			purpose,
			&upload.Registration{
				Policy:  policy,
				Handler: handler,
			},
		)
	}

	return registry, nil
}
