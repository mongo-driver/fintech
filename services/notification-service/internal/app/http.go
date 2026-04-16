package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	svc *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

type manualRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/send", h.send)
}

func (h *HTTPHandler) send(c *gin.Context) {
	var req manualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SendManual(c.Request.Context(), req.UserID, req.Subject, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}
