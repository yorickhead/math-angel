package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/repository"
	"github.com/osamikoyo/math-angel/internal/service"
)

func (h *Handler) DecLike(c *echo.Context) error {
	id := c.Param("id")

	if err := h.service.DecLike(c.Request().Context(), id); err != nil {
		switch err {
		case repository.ErrNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case service.ErrBadUID:
			return c.String(http.StatusBadRequest, err.Error())
		default:
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	return c.String(http.StatusOK, "success")
}
