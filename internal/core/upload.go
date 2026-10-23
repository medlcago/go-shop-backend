package core

import (
	"go-shop-backend/internal/upload"
	uploadHandlers "go-shop-backend/internal/upload/handlers"
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

func newUploadHandlers(container *Container) map[upload.Purpose]upload.Handler {
	handlers := map[upload.Purpose]upload.Handler{
		upload.ProductImagePurpose: uploadHandlers.NewProductImageHandler(
			container.ProductRepo(),
			container.UploadRepo(),
			container.Config().Upload.MaxImagesPerProduct,
		),
	}

	return handlers
}

func newUploadRegistry(container *Container) upload.Registry {
	handlers := newUploadHandlers(container)

	registry := upload.NewRegistry()

	registry.Register(
		upload.ProductImagePurpose,
		&upload.Registration{
			Policy:  productImagePolicy,
			Handler: handlers[upload.ProductImagePurpose],
		},
	)

	return registry
}
