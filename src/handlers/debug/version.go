package debug

import (
	"net/http"
	"setting"

	"modules/responser"

	"github.com/labstack/echo"
)

// Version Version
func Version(c echo.Context) error {
	return responser.R(c, http.StatusOK, setting.Conf.Version)
}
