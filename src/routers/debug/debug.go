package debug

import (
	"routers"

	h "handlers/debug"

	"github.com/labstack/echo"
)

var _ routers.RouterRegister = debugRouters

func debugRouters(e *echo.Echo) {
	r := e.Group("/debug")
	r.GET("/version", h.Version)

}

func init() {
	routers.Register(debugRouters)
}
