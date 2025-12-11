package auth

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	authService service.AuthService
}

func NewHandler(authService service.AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

// Login godoc
// @Summary Login
// @Description Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UserLoginRequest true "Request body for login"
// @Success 200 {object} response.Response[dto.LoginResponse]
// @Failure 400 {object} response.Response[any]
// @Failure 401 {object} response.Response[any]
// @Failure 500 {object} response.Response[any]
// @Router /auth/login [post]
func (h *Handler) Login(ctx fiber.Ctx) error {
	var req dto.UserLoginRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.authService.Login(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Register godoc
// @Summary Register
// @Description Register
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.UserRegisterRequest true "Request body for registration"
// @Success 201 {object} response.Response[dto.RegisterResponse]
// @Failure 400 {object} response.Response[any]
// @Failure 409 {object} response.Response[any] "The user already exists"
// @Failure 500 {object} response.Response[any]
// @Router /auth/register [post]
func (h *Handler) Register(ctx fiber.Ctx) error {
	var req dto.UserRegisterRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.authService.Register(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}
