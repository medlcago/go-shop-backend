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
		return fiber.ErrBadRequest
	}

	resp, err := h.uploadService.SignURL(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}
