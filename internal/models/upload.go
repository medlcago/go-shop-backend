package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadMediaType string

const (
	UploadMediaTypeDefault UploadMediaType = "default"
)

type UploadVariant string

const (
	UploadVariantOriginal UploadVariant = "original"
)

type EntityType string

const (
	EntityTypeProduct EntityType = "products"
)

type Upload struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ObjectKey string `gorm:"type:varchar(1024);not null;uniqueIndex:idx_uploads_object_key_unique,where:deleted_at IS NULL"`

	EntityType EntityType `gorm:"type:varchar(255);not null;index:idx_uploads_entity_created_at,priority:1,where:deleted_at IS NULL;index:idx_uploads_entity_media_type,priority:1,where:deleted_at IS NULL"`
	EntityID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_uploads_entity_created_at,priority:2,where:deleted_at IS NULL;index:idx_uploads_entity_media_type,priority:2,where:deleted_at IS NULL"`

	FileSize    int64   `gorm:"not null"`
	ContentType *string `gorm:"type:varchar(255)"`

	MediaType UploadMediaType `gorm:"type:varchar(100);not null;default:'default';index:idx_uploads_entity_media_type,priority:3,where:deleted_at IS NULL"`
	Variant   UploadVariant   `gorm:"type:varchar(50);not null;default:'original'"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now();not null;index:idx_uploads_entity_created_at,priority:3,sort:desc,where:deleted_at IS NULL"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`

	URL string `gorm:"-"`
}

func (u *Upload) GetObjectKey() string {
	return u.ObjectKey
}

func (u *Upload) SetURL(url string) {
	u.URL = url
}
