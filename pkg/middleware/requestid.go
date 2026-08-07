package middleware

import (
	"go-shop-backend/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func RequestID(ctx fiber.Ctx) error {
	requestID, err := uuid.Parse(fiber.GetReqHeader[string](ctx, fiber.HeaderXRequestID))
	if err != nil {
		requestID = uuid.New()
	}

	rid := requestID.String()

	ctx.Set(fiber.HeaderXRequestID, rid)

	reqCtx := logger.WithRequestID(ctx.Context(), rid)
	ctx.SetContext(reqCtx)

	return ctx.Next()
}
