package repository

import (
	"context"
	"go-shop-backend/pkg/paging"
	"go-shop-backend/pkg/transaction"

	"github.com/huandu/go-sqlbuilder"
)

func PaginatedQuery[T any](
	ctx context.Context,
	db transaction.Queryer,
	countBuilder, selectBuilder *sqlbuilder.SelectBuilder,
	limit, offset int,
) ([]*T, int64, error) {
	countQuery, countArgs := countBuilder.Build()

	var total int64
	if err := db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	selectBuilder.WhereClause = countBuilder.WhereClause

	pagination := paging.New(limit, offset)
	selectBuilder.Limit(pagination.Limit).Offset(pagination.Offset)

	query, args := selectBuilder.Build()

	var dest []*T
	if err := db.SelectContext(ctx, &dest, query, args...); err != nil {
		return nil, 0, HandleSQLError(err)
	}

	return dest, total, nil
}
