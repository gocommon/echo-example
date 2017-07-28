package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"routers"
	"setting"

	"github.com/labstack/echo/middleware"

	"github.com/labstack/echo"
)

var configPath string

func init() {
	pwd, _ := os.Getwd()
	flag.StringVar(&configPath, "c", filepath.Join(pwd, "./src/cmd/main/app.toml"), "-c /path/to/app.toml config gile")
}

// 一些初始化工作
func bootstrap() {
	if err := setting.InitConf(configPath); err != nil {
		panic(err)
	}
	// 版本号
	setting.Conf.Version = VER

	log.Println(setting.Conf)
}

func main() {
	flag.Parse()

	bootstrap()

	e := echo.New()
	e.Debug = setting.Conf.Echo.Debug
	e.HideBanner = setting.Conf.Echo.HideBanner

	// 中间件 访问日志，写文件
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 统计错误处理
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var (
			code = http.StatusInternalServerError
			msg  interface{}
		)

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = he.Message
		} else if e.Debug {
			msg = err.Error()
		} else {
			msg = http.StatusText(code)
		}
		if _, ok := msg.(string); ok {
			msg = echo.Map{"message": msg}
		}

		if !c.Response().Committed {
			if c.Request().Method == echo.HEAD { // Issue #608
				if err := c.NoContent(code); err != nil {
					goto ERROR
				}
			} else {
				if err := c.JSON(code, msg); err != nil {
					goto ERROR
				}
			}
		}
	ERROR:
		e.Logger.Error(err)
	}

	// 注册路由
	routers.InitRouters(e)

	log.Println(setting.Conf.Version)

	e.Start(setting.Conf.Echo.Listen)
}
