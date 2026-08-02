package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Health is a handler for checking the health of the service.
//
//	@Summary		Check health
//	@Description	Chech health
//	@Tags			API
//	@Produce		json
//	@Success		200		{string}	OK
//	@Router			/health [get]
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}
