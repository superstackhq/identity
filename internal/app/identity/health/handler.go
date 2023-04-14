package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superstackhq/common/api"
)

type Handler struct {
	router *gin.Engine
}

func NewHandler(router *gin.Engine) *Handler {
	return &Handler{
		router: router,
	}
}

func (h *Handler) Register() {
	h.router.GET("/health", h.healthCheck)
}

func (h *Handler) healthCheck(c *gin.Context) {
	api.Success(c, http.StatusOK, "health check passed successfully")
}
