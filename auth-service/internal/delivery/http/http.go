package http

import (
	"encoding/json"
	"net/http"

	"github.com/MiRRoRise/auth-service/internal/delivery/dto"
	"github.com/MiRRoRise/auth-service/internal/domain"
	"github.com/MiRRoRise/auth-service/internal/usecase"
	"github.com/MiRRoRise/auth-service/pkg/jwt"
	"github.com/MiRRoRise/auth-service/pkg/logger"
)

type Handler struct {
	userUseCase  usecase.UserUseCase
	tokenManager jwt.TokenManager
}

func NewHandler(userUseCase usecase.UserUseCase, tokenManager jwt.TokenManager) *Handler {
	return &Handler{
		userUseCase:  userUseCase,
		tokenManager: tokenManager,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	JSON(w, map[string]string{"status": "ok"}, http.StatusOK)
}

// Register godoc
// @Summary Register new user
// @Description Create new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "error decode json", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		Error(w, "error email or password required", http.StatusBadRequest)
		return
	}

	user, err := h.userUseCase.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserExists:
			Error(w, "user already exists", http.StatusConflict)
		default:
			Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	JSON(w, resp, http.StatusCreated)
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200  {object}  dto.LoginResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "error decode req", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		Error(w, "email or password required", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.userUseCase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound, domain.ErrInvalidCredentials:
			Error(w, "invalid credentials", http.StatusUnauthorized)
		default:
			Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	JSON(w, resp, http.StatusOK)
}

// Me godoc
// @Summary      Get current user profile
// @Description  Get detailed info about authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserResponse
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := GetIDFromContext(r.Context())
	if err != nil {
		Error(w, "user unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userUseCase.GetUserByID(r.Context(), userID)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			Error(w, "user not found", http.StatusNotFound)
		default:
			Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	JSON(w, resp, http.StatusOK)
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Get a new pair of access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshRequest true "Refresh token data"
// @Success      200  {object}  dto.RefreshResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "error decode refresh token", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		Error(w, "refresh token is required", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.userUseCase.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			Error(w, "user not found", http.StatusNotFound)
		case domain.ErrInvalidToken:
			Error(w, "invalid token", http.StatusBadRequest)
		default:
			Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	JSON(w, resp, http.StatusOK)
}

// Logout godoc
// @Summary      Logout user
// @Description  Client-side logout only (no server-side token revocation)
// @Tags         auth
// @Success      204  "No Content"
// @Router       /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func JSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.New().Error("error encode json", err)
		}
	}
}

func Error(w http.ResponseWriter, message string, status int) {
	JSON(w, map[string]string{"error": message}, status)
}
