package httpx

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error string `json:"error"`
}

func BadRequest(c *gin.Context, err error) {
	c.JSON(400, ErrorResponse{Error: err.Error()})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(401, ErrorResponse{Error: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(404, ErrorResponse{Error: msg})
}

func Internal(c *gin.Context, err error) {
	c.JSON(500, ErrorResponse{Error: err.Error()})
}
