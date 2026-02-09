package user

import (
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	userService service.UserService
}

func NewHandler(userService service.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// GetMe godoc
//
//	@Summary		Get Me
//	@Description	Get Me
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[internal_dto.UserResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/users/me [get]
func (h *Handler) GetMe(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperrors.ErrInvalidCredentials
	}

	resp, err := h.userService.GetUserByID(ctx, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
