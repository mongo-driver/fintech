package app

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/fintech-backend/services/user-service/internal/repository"
	"github.com/example/fintech-backend/shared/httpx"
)

type HTTPHandler struct {
	svc *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/", h.createUser)
	r.GET("/", h.listUsers)
	r.GET("/:id", h.getUser)
	r.PUT("/:id", h.updateUser)
	r.DELETE("/:id", h.deleteUser)
}

func (h *HTTPHandler) createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	user, err := h.svc.CreateUser(c.Request.Context(), req)
	if err != nil {
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *HTTPHandler) getUser(c *gin.Context) {
	user, err := h.svc.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		if err == repository.ErrNotFound {
			httpx.NotFound(c, "user not found")
			return
		}
		httpx.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *HTTPHandler) updateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	user, err := h.svc.UpdateUser(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if err == repository.ErrNotFound {
			httpx.NotFound(c, "user not found")
			return
		}
		httpx.BadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *HTTPHandler) deleteUser(c *gin.Context) {
	if err := h.svc.DeleteUser(c.Request.Context(), c.Param("id")); err != nil {
		if err == repository.ErrNotFound {
			httpx.NotFound(c, "user not found")
			return
		}
		httpx.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) listUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	users, err := h.svc.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		httpx.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}
