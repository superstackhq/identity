package organization

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superstackhq/common/api"
	"github.com/superstackhq/identity/internal/app/identity/authentication"
)

type Handler struct {
	router        *gin.Engine
	authenticator *authentication.Authenticator
	manager       *Manager
}

func NewHandler(router *gin.Engine, authenticator *authentication.Authenticator, manager *Manager) *Handler {
	return &Handler{
		router:        router,
		authenticator: authenticator,
		manager:       manager,
	}
}

func (h *Handler) Register() {
	h.router.GET("/api/v1/organization", h.get)
}

func (h *Handler) get(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	au, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	org, err := h.manager.Get(ctx, au.OrganizationID)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, org)
}
