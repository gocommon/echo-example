package debug

import (
	"routers"

	h "handlers/debug"

	"github.com/labstack/echo"
)

// 实现注册路由方法
var _ routers.RouterRegister = debugRouters

func debugRouters(e *echo.Echo) {
	r := e.Group("/debug")
	r.GET("/version", h.Version)

}

// 注册到路由表
func init() {
	routers.Register(debugRouters)
}
