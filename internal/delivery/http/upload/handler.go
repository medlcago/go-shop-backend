package upload

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	uploadManager upload.Manager
}

func NewHandler(uploadManager upload.Manager) *Handler {
	return &Handler{
		uploadManager: uploadManager,
	}
}

func (h *Handler) SignURL(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.GeneratePresignedURLRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.uploadManager.SignURL(ctx.Context(), req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) Attach(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.AttachFileRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.uploadManager.Attach(ctx.Context(), req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
