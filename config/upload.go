package config

import "time"

type Upload struct {
	MaxFileSize     int64         `env:"MAX_FILE_SIZE" env-default:"5242880"` // 5MB
	PresignedURLTTL time.Duration `env:"PRESIGNED_URL_TTL" env-default:"20m"`

	MaxImagesPerProduct int64 `env:"MAX_IMAGES_PER_PRODUCT" env-default:"10"`
}
