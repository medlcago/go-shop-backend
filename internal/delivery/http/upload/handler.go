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

// SignURL godoc
//
//	@Summary		Generate a presigned upload URL
//	@Description	Generates a presigned URL (and optional form-data) for direct upload to an S3-compatible object storage. Requires authentication.
//	@Tags			Uploads
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.GeneratePresignedURLRequest	true	"Presigned URL generation parameters"
//	@Success		200		{object}	response.Response[dto.GeneratePresignedURLResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		403		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/uploads/signurl [post]
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

// Attach godoc
//
//	@Summary		Attach an uploaded file to an entity
//	@Description	Persists a successfully uploaded file and links it to an entity (e.g. product, user avatar, etc.). Requires authentication.
//	@Tags			Uploads
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.AttachFileRequest	true	"File attachment parameters"
//	@Success		200		{object}	response.Response[dto.AttachFileResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		403		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/uploads/attach [post]
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
