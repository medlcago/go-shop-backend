package user

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	userService service.UserService
}

func NewHandler(userService service.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// Login godoc
//
//	@Summary		Login
//	@Description	Login with email and password
//	@Tags			Auth
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

	resp, err := h.userService.Login(ctx.Context(), req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Register godoc
//
//	@Summary		Register
//	@Description	Register a new user account
//	@Tags			Auth
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

	resp, err := h.userService.Register(ctx.Context(), req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}

// Setup2FA godoc
//
//	@Summary		Setup 2FA
//	@Description	Initialize two-factor authentication setup for the authenticated user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.Setup2FAResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		409	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/setup-2fa [post]
func (h *Handler) Setup2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.Setup2FA(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Confirm2FA godoc
//
//	@Summary		Confirm 2FA
//	@Description	Confirm and enable two-factor authentication with the provided code
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Confirm2FARequest		true	"Request body with 2FA confirmation code"
//	@Success		200		{object}	response.Response[string]	"2FA successfully enabled"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/confirm-2fa [post]
func (h *Handler) Confirm2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.Confirm2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.userService.Confirm2FA(ctx.Context(), *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// Disable2FA godoc
//
//	@Summary		Disable 2FA
//	@Description	Disable two-factor authentication for the authenticated user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Disable2FARequest		true	"Request body with password confirmation"
//	@Success		200		{object}	response.Response[string]	"2FA successfully disabled"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/disable-2fa [post]
func (h *Handler) Disable2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.Disable2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.userService.Disable2FA(ctx.Context(), *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// RefreshToken godoc
//
//	@Summary		Refresh Token
//	@Description	Refresh access token using a valid refresh token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.UserTokenResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/refresh [post]
func (h *Handler) RefreshToken(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)

	resp, err := h.userService.RefreshToken(ctx.Context(), userCtx.Token)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// GetMe godoc
//
//	@Summary		Get Me
//	@Description	Get current authenticated user profile
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.UserResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/users/me [get]
func (h *Handler) GetMe(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.GetUserByID(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// SendEmailConfirmationCode godoc
//
//	@Summary		Send email confirmation code
//	@Description	Send a confirmation code to the user's email address for verification
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.SendEmailConfirmationResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		429	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/users/me/send-email-confirmation [post]
func (h *Handler) SendEmailConfirmationCode(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.SendEmailConfirmationCode(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// ConfirmEmail godoc
//
//	@Summary		Confirm email address
//	@Description	Confirm user's email address using the confirmation code sent to their email
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.ConfirmEmailRequest	true	"Request body with confirmation code"
//	@Success		200		{object}	response.Response[dto.ConfirmEmailResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		409		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]	"Internal server error"
//	@Router			/users/me/confirm-email [post]
func (h *Handler) ConfirmEmail(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.ConfirmEmailRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.userService.ConfirmEmail(ctx.Context(), *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// ChangePassword godoc
//
//	@Summary		Change password
//	@Description	Change user password
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.ChangePasswordRequest	true	"Request body with current and new password"
//	@Success		200		{object}	response.Response[string]	"OK"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/users/me/change-password [post]
func (h *Handler) ChangePassword(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.ChangePasswordRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.userService.ChangePassword(ctx.Context(), *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// BeginPasskeyRegistration godoc
//
//	@Summary		Begin Passkey Registration
//	@Description	Start WebAuthn passkey registration ceremony for authenticated user
//	@Tags			Auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.BeginPasskeyRegistrationResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/passkeys/register/begin [post]
func (h *Handler) BeginPasskeyRegistration(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.BeginPasskeyRegistration(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// FinishPasskeyRegistration godoc
//
//	@Summary		Finish Passkey Registration
//	@Description	Complete WebAuthn passkey registration ceremony using credential creation response
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Passkey-Session-ID	header		string	true	"Passkey Session ID obtained during registration start"
//	@Param			request				body		object	true	"WebAuthn Credential Creation Response JSON"
//	@Success		200					{object}	response.Response[string]
//	@Failure		400					{object}	response.Response[any]
//	@Failure		401					{object}	response.Response[any]
//	@Failure		500					{object}	response.Response[any]
//	@Router			/auth/passkeys/register/finish [post]
func (h *Handler) FinishPasskeyRegistration(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	sid, err := getPasskeySessionID(ctx)
	if err != nil {
		return err
	}

	if err := h.userService.FinishPasskeyRegistration(ctx.Context(), *userCtx.UserID, sid, ctx.Body()); err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// BeginPasskeyLogin godoc
//
//	@Summary		Begin Passkey Login
//	@Description	Start WebAuthn passkey authentication ceremony
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	response.Response[dto.BeginPasskeyDiscoverableLoginResponse]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/passkeys/login/begin [post]
func (h *Handler) BeginPasskeyLogin(ctx fiber.Ctx) error {
	resp, err := h.userService.BeginPasskeyLogin(ctx.Context())
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// FinishPasskeyLogin godoc
//
//	@Summary		Finish Passkey Login
//	@Description	Complete WebAuthn passkey authentication ceremony and issue token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			Passkey-Session-ID	header		string	true	"Passkey Session ID obtained during login start"
//	@Param			request				body		object	true	"WebAuthn Credential Assertion Response JSON"
//	@Success		200					{object}	response.Response[dto.UserTokenResponse]
//	@Failure		400					{object}	response.Response[any]
//	@Failure		401					{object}	response.Response[any]
//	@Failure		500					{object}	response.Response[any]
//	@Router			/auth/passkeys/login/finish [post]
func (h *Handler) FinishPasskeyLogin(ctx fiber.Ctx) error {
	sid, err := getPasskeySessionID(ctx)
	if err != nil {
		return err
	}

	resp, err := h.userService.FinishPasskeyLogin(ctx.Context(), sid, ctx.Body())
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// GetUserPasskeys godoc
//
//	@Summary		Get User Passkeys
//	@Description	Get list of all passkeys registered for the current user
//	@Tags			Auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[[]dto.PasskeyResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/passkeys [get]
func (h *Handler) GetUserPasskeys(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.GetUserPasskeys(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// UpdatePasskeyName godoc
//
//	@Summary		Update Passkey Name
//	@Description	Update custom display name of a registered passkey
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Passkey UUID"	format(uuid)
//	@Param			request	body		dto.UpdatePasskeyNameRequest	true	"Request body to update passkey name"
//	@Success		200		{object}	response.Response[string]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/auth/passkeys/{id}/name [put]
func (h *Handler) UpdatePasskeyName(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	id := uuid.MustParse(ctx.Params("id"))

	var req dto.UpdatePasskeyNameRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	if err := h.userService.UpdatePasskeyName(ctx.Context(), id, *userCtx.UserID, req); err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// DeletePasskey godoc
//
//	@Summary		Delete Passkey
//	@Description	Remove a passkey registered by the current user
//	@Tags			Auth
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Passkey UUID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/auth/passkeys/{id} [delete]
func (h *Handler) DeletePasskey(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	id := uuid.MustParse(ctx.Params("id"))

	if err := h.userService.DeletePasskey(ctx.Context(), id, *userCtx.UserID); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func getPasskeySessionID(ctx fiber.Ctx) (string, error) {
	sid := fiber.GetReqHeader[string](ctx, "Passkey-Session-ID")
	if sid == "" {
		return "", apperror.ErrInvalidPasskeySessionID
	}

	return sid, nil
}
