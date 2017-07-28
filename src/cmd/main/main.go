package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"routers"
	"setting"

	"modules/validator"

	"github.com/gocommon/rotatefile"
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

	// 参数验证
	e.Validator = validator.New()

	///////////////// 中间件 ////////////////
	///									////

	// 访问日志，写文件
	if setting.Conf.Echo.AccessLog {
		loggerConfig := middleware.DefaultLoggerConfig

		if setting.Conf.Echo.AccessLogFile {
			f, err := rotatefile.NewWriter(rotatefile.Options{Filename: setting.Conf.Echo.AccessLogFilePath})
			defer f.Close()
			loggerConfig.Output = f
			if err != nil {
				panic(err)
			}
		}

		e.Use(middleware.LoggerWithConfig(loggerConfig))
	}

	e.Use(middleware.Recover())

	////								////
	///////////////// 中间件 ////////////////

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

	// @todo tls
	e.Start(setting.Conf.Echo.Listen)
}
