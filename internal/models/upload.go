package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Upload struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ObjectKey string `gorm:"type:varchar(1024);not null;
		uniqueIndex:idx_uploads_object_key_unique,where:deleted_at IS NULL"`

	EntityType string `gorm:"type:varchar(255);not null;
		index:idx_uploads_entity_created_at,priority:1,where:deleted_at IS NULL;
		uniqueIndex:idx_uploads_is_main_unique,priority:1,where:is_main = true AND deleted_at IS NULL"`
	EntityID uuid.UUID `gorm:"type:uuid;not null;
		index:idx_uploads_entity_created_at,priority:2,where:deleted_at IS NULL;
		uniqueIndex:idx_uploads_is_main_unique,priority:2,where:is_main = true AND deleted_at IS NULL"`

	FileSize    int64   `gorm:"not null"`
	ContentType *string `gorm:"type:varchar(255)"`

	IsMain bool `gorm:"not null;default:false"`

	CreatedAt time.Time `gorm:"index:idx_uploads_entity_created_at,priority:3,sort:desc,where:deleted_at IS NULL"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
