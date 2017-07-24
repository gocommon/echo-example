package routers

import "github.com/labstack/echo"

// RouterRegister RouterRegister
type RouterRegister func(*echo.Echo)

var routers = []RouterRegister{}

// Register Register
func Register(r RouterRegister) {
	routers = append(routers, r)
}

// InitRouters InitRouters
func InitRouters(e *echo.Echo) {
	for i := range routers {
		routers[i](e)
	}
}
