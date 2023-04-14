package user

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superstackhq/common/api"
	"github.com/superstackhq/identity/internal/app/identity/authentication"
	"github.com/superstackhq/identity/pkg/actor"
	"github.com/superstackhq/identity/pkg/user"
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
	h.router.POST("/api/v1/accounts/signup", h.signUp)
	h.router.POST("/api/v1/accounts/authenticate", h.authenticate)

	h.router.GET("/api/v1/users/me", h.get)
	h.router.PUT("/api/v1/users/me/password", h.changePassword)

	h.router.POST("/api/v1/users", h.add)
	h.router.DELETE("/api/v1/users/:userID", h.delete)
	h.router.GET("/api/v1/users", h.list)
	h.router.GET("/api/v1/users/:userID", h.getByOrganization)
	h.router.PUT("/api/v1/users/:userID/admin", h.changeAdmin)
	h.router.PUT("/api/v1/users/:userID/password", h.resetPassword)
}

func (h *Handler) signUp(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	var request user.SignUpRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		api.Error(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.manager.SignUp(ctx, &request)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) authenticate(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	var request user.AuthenticationRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		api.Error(c, http.StatusBadRequest, err)
		return
	}

	response, err := h.manager.Authenticate(ctx, &request)
	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) get(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	if a.ActorType != actor.TypeUser {
		api.ErrorMessage(c, http.StatusForbidden, "not allowed")
		return
	}

	u, err := h.manager.Get(ctx, a.ActorID)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) changePassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	if a.ActorType != actor.TypeUser {
		api.ErrorMessage(c, http.StatusForbidden, "not allowed")
		return
	}

	var request user.PasswordChangeRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	u, err := h.manager.ChangePassword(ctx, a.ActorID, &request)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) add(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	if !a.HasFullAccess {
		api.ErrorMessage(c, http.StatusForbidden, "not allowed")
		return
	}

	var request user.AdditionRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		api.Error(c, http.StatusBadRequest, err)
		return
	}

	p, err := h.manager.Add(ctx, &request, a)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *Handler) delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	if !a.HasFullAccess {
		api.ErrorMessage(c, http.StatusForbidden, "not allowed")
		return
	}

	userID, ok := c.Params.Get("userID")

	if !ok {
		api.ErrorMessage(c, http.StatusBadRequest, "user id is required")
		return
	}

	u, err := h.manager.Delete(ctx, userID, a.OrganizationID)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) list(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	page, size := api.Page(c)

	users, err := h.manager.List(ctx, a.OrganizationID, page, size)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) getByOrganization(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	userID, ok := c.Params.Get("userID")

	if !ok {
		api.ErrorMessage(c, http.StatusBadRequest, "user id is required")
		return
	}

	user, err := h.manager.GetByOrganization(ctx, userID, a.OrganizationID)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) resetPassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	if !a.HasFullAccess {
		api.ErrorMessage(c, http.StatusForbidden, "not allowed")
		return
	}

	userID, ok := c.Params.Get("userID")

	if !ok {
		api.ErrorMessage(c, http.StatusBadRequest, "user id is required")
		return
	}

	p, err := h.manager.ResetPassword(ctx, userID, a.OrganizationID)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) changeAdmin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	a, err := h.authenticator.ValidateContext(c, ctx)

	if err != nil {
		api.Error(c, http.StatusUnauthorized, err)
		return
	}

	if !a.HasFullAccess {
		api.ErrorMessage(c, http.StatusForbidden, "not allowed")
		return
	}

	userID, ok := c.Params.Get("userID")

	if !ok {
		api.ErrorMessage(c, http.StatusBadRequest, "user id is required")
		return
	}

	var request user.AdminChangeRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		api.Error(c, http.StatusBadRequest, err)
		return
	}

	u, err := h.manager.ChangeAdmin(ctx, userID, request, a.OrganizationID)

	if err != nil {
		api.Error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, u)
}
