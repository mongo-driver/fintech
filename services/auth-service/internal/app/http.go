package app

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/fintech-backend/shared/httpx"
	"github.com/example/fintech-backend/shared/security"
)

type HTTPHandler struct {
	svc *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/register", h.register)
	r.POST("/login", h.login)
	r.POST("/validate", h.validateToken)
	r.POST("/oauth/token", h.oauthToken)
}

func (h *HTTPHandler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	resp, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HTTPHandler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.Unauthorized(c, err.Error())
			return
		}
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

type validateRequest struct {
	Token string `json:"token"`
}

func (h *HTTPHandler) validateToken(c *gin.Context) {
	var req validateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	claims, err := h.svc.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		httpx.Unauthorized(c, "invalid token")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"exp":     claims.ExpiresAt.Time,
	})
}

func (h *HTTPHandler) oauthToken(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "oauth2 token endpoint scaffolded; integrate provider for production federation.",
	})
}

type AuthMiddleware struct {
	manager *security.JWTManager
}
