package dto

import "github.com/google/uuid"

type CategoryResponse struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active"`
	SortOrder   int        `json:"sort_order"`
}

type ListCategoryRequest struct {
	ID uuid.UUID `json:"-"`

	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}
