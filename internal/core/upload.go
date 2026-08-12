package core

import (
	"go-shop-backend/internal/service"
	"go-shop-backend/internal/upload"
)

var (
	productImagePolicy = upload.Policy{
		MinSize: 5 << 10, // 5 KB
		MaxSize: 5 << 20, // 5 MB
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

func NewUploadRegistry() upload.Registry {
	registry := upload.NewRegistry()
	registry.Register(
		service.ProductImageType,
		productImagePolicy,
	)

	return registry
}
