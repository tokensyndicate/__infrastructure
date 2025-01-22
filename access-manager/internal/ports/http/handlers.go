// internal/ports/http/handlers.go
package http

import (
	"net/http"

	"access-manager/internal/domain"
	"access-manager/internal/service"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	manager *service.AccessManager
}

func NewHandlers(manager *service.AccessManager) *Handlers {
	return &Handlers{manager: manager}
}

func (h *Handlers) RegisterInstance(c echo.Context) error {
	var instance domain.Instance
	if err := c.Bind(&instance); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.manager.RegisterInstance(c.Request().Context(), instance); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, instance)
}

func (h *Handlers) GetInstance(c echo.Context) error {
	id := c.Param("id")
	instance, err := h.manager.GetInstance(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, instance)
}

func (h *Handlers) AssignCredentials(c echo.Context) error {
	var creds domain.Credentials
	if err := c.Bind(&creds); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ip, err := h.manager.AssignCredentials(c.Request().Context(), creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"external_ip": ip})
}

func (h *Handlers) UpdateInstanceMetrics(c echo.Context) error {
	var metrics domain.InstanceMetrics
	if err := c.Bind(&metrics); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.manager.UpdateInstanceMetrics(c.Request().Context(), metrics); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handlers) GetInstanceBindings(c echo.Context) error {
	id := c.Param("id")
	bindings, err := h.manager.GetInstanceBindings(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, bindings)
}

func (h *Handlers) GetInstanceCredentials(c echo.Context) error {
	instanceID := c.Param("id")

	bindings, err := h.manager.GetInstanceBindings(c.Request().Context(), instanceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Получаем все credentials для найденных bindings
	var credentials []domain.Credentials
	for _, binding := range bindings {
		creds, err := h.manager.GetCredentials(c.Request().Context(), binding.CredentialsID)
		if err != nil {
			continue // Пропускаем проблемные credentials
		}
		credentials = append(credentials, *creds)
	}

	return c.JSON(http.StatusOK, credentials)
}
