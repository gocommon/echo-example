package main

import (
	"log"
	"routers"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()

	// 统计错误处理

	// 注册路由
	routers.InitRouters(e)

	log.Println(VER)

	e.Start(":8899")
}
