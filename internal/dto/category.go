package dto

import "github.com/google/uuid"

type ProductCategoryResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	ParentID string `json:"parent_id"`
}

type ListCategoryRequest struct {
	ID uuid.UUID `json:"-"`

	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}
