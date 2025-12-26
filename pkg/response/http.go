package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func JSON(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Success: status >= 200 && status < 300,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string, errs interface{}) {
	c.JSON(status, Response{
		Success: false,
		Message: message,
		Errors:  errs,
	})
}

func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, "Success", data)
}
