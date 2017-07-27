package debug

import (
	"net/http"
	"setting"

	"github.com/labstack/echo"
)

// Version Version
func Version(c echo.Context) error {
	c.Logger().Debugf("ver:%s", setting.Conf.Version)
	return c.String(http.StatusOK, setting.Conf.Version)
}
