package routers

import "github.com/labstack/echo"

// RouterRegister 定义注册路由方法
type RouterRegister func(*echo.Echo)

var routers = []RouterRegister{}

// Register 注册路由方法，供模块添加路由
func Register(r RouterRegister) {
	routers = append(routers, r)
}

// InitRouters 初始化路由，echo.Start 前执行
func InitRouters(e *echo.Echo) {
	for i := range routers {
		routers[i](e)
	}
}
