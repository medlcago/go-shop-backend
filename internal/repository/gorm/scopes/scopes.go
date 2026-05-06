package scopes

import (
	"go-shop-backend/pkg/paging"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Paginate(limit, offset int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		pagination := paging.New(limit, offset)
		return db.Limit(pagination.Limit).Offset(pagination.Offset)
	}
}

func ProductOrderBy(orderBy string, orderDesc bool) func(db *gorm.DB) *gorm.DB {
	allowedOrderBy := map[string]struct{}{
		"id":         {},
		"created_at": {},
		"price":      {},
	}

	return func(db *gorm.DB) *gorm.DB {
		order := "products.created_at"
		if _, ok := allowedOrderBy[orderBy]; ok {
			order = "products." + orderBy
			if orderDesc {
				order += " DESC"
			}
		}

		return db.Order(order)
	}
}

func ProductWithCategory(categoryID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if categoryID == uuid.Nil {
			return db
		}

		return db.Where(`
			EXISTS (
				SELECT 1
				FROM product_categories pc
				JOIN categories c ON c.id = pc.category_id
				WHERE pc.product_id = products.id
				  AND pc.category_id = ?
				  AND c.is_active = true
			)
		`, categoryID)
	}
}

func AvailableProducts() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ? AND stock > ?", true, 0)
	}
}

func ProductPriceBetween(min int64, max int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if min <= 0 {
			min = 0
		}

		if max <= 0 {
			max = math.MaxInt64
		}

		return db.Where("price BETWEEN ? AND ?", min, max)
	}
}

func ProductWithRelations() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Images").Preload("Categories", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "slug").Where("is_active = ?", true)
		})
	}
}

func OrderStatus(status string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == "" {
			return db
		}

		return db.Where("status = ?", status)
	}
}

func OrderOwner(userID *uuid.UUID, sessionID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if userID != nil {
			return db.Where(
				"(user_id = ? OR session_id = ?)",
				*userID,
				sessionID,
			)
		}

		return db.Where("session_id = ?", sessionID)
	}
}

func OrderWithRelations() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("order_items.created_at ASC").
				Preload("Product", func(tx *gorm.DB) *gorm.DB {
					return tx.Select("id").
						Preload("Images", func(tx *gorm.DB) *gorm.DB {
							return tx.Select("id", "object_key", "entity_id", "entity_type")
						})
				})
		})
	}
}

func CategoryWithParentID(parentID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if parentID == uuid.Nil {
			return db.Where("parent_id IS NULL")
		}

		return db.Where("parent_id = ?", parentID)
	}
}

func WishlistWithRelations() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("priority DESC, created_at DESC")
		}).Preload("Items.Product")
	}
}
