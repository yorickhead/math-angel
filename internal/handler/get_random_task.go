package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

func (h *Handler) GetRandomTask(c *echo.Context) error {
	levelStr := c.Param("level")
	taskType := c.Param("type")

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "level must be number")
	}

	task, err := h.service.GetRandomTask(c.Request().Context(), taskType, uint(level))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, task)
}
