package upload

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	uploadService service.UploadService
}

func NewHandler(uploadService service.UploadService) *Handler {
	return &Handler{
		uploadService: uploadService,
	}
}

// SignURL godoc
//
//	@Summary		Generate signed URL for upload
//	@Description	Create a pre-signed URL for uploading files to storage
//	@Tags			uploads
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.SignURLRequest	true	"Sign URL request parameters"
//	@Success		201		{object}	response.Response[dto.SignURLResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		403		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/uploads/signurl [post]
func (h *Handler) SignURL(ctx fiber.Ctx) error {
	var req dto.SignURLRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.uploadService.SignURL(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}

// Save godoc
//
//	@Summary		Attach uploaded file to entity
//	@Description	Bind a previously uploaded file (via signed URL) to a specific entity after successful upload
//	@Tags			uploads
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.UploadRequest	true	"Upload save request parameters"
//	@Success		200		{object}	response.Response[dto.UploadResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		403		{object}	response.Response[any]
//	@Failure		409		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/uploads/save [post]
func (h *Handler) Save(ctx fiber.Ctx) error {
	var req dto.UploadRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.uploadService.Save(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
