package auth

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
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
//
//	@Summary		Login
//	@Description	Login
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserLoginRequest	true	"Request body for login"
//	@Success		200		{object}	response.Response[dto.UserTokenResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/login [post]
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

// LoginWith2FA godoc
//
//	@Summary		Login with 2FA
//	@Description	Authenticate user with two-factor authentication code
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.Verify2FARequest	true	"Request body with 2FA code"
//	@Success		200		{object}	response.Response[dto.UserTokenResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/login/2fa [post]
func (h *Handler) LoginWith2FA(ctx fiber.Ctx) error {
	var req dto.Verify2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.authService.Verify2FA(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Register godoc
//
//	@Summary		Register
//	@Description	Register
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserRegisterRequest	true	"Request body for registration"
//	@Success		201		{object}	response.Response[dto.UserTokenResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		409		{object}	response.Response[any]	"The user already exists"
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/register [post]
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

// Setup2FA godoc
//
//	@Summary		Setup 2FA
//	@Description	Initialize two-factor authentication setup for the authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.Setup2FAResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		409	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/2fa/setup [post]
func (h *Handler) Setup2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.authService.Setup2FA(ctx, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Confirm2FA godoc
//
//	@Summary		Confirm 2FA
//	@Description	Confirm and enable two-factor authentication with the provided code
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Confirm2FARequest		true	"Request body with 2FA confirmation code"
//	@Success		200		{object}	response.Response[string]	"2FA successfully enabled"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/2fa/confirm [post]
func (h *Handler) Confirm2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.Confirm2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.authService.Confirm2FA(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// Disable2FA godoc
//
//	@Summary		Disable 2FA
//	@Description	Disable two-factor authentication for the authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Disable2FARequest		true	"Request body with password confirmation"
//	@Success		200		{object}	response.Response[string]	"2FA successfully disabled"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/2fa/disable [post]
func (h *Handler) Disable2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.Disable2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.authService.Disable2FA(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}
