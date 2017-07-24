package debug

import (
	"net/http"

	"github.com/labstack/echo"
)

// Version Version
func Version(c echo.Context) error {
	return c.String(http.StatusOK, "version")
}
