package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"routers"
	"setting"
	"time"

	"modules/validator"
	"modules/zerolog"

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

	zerolog.InitLog(setting.Conf.ZeroLogs, zerolog.Timestamp(), zerolog.Version(setting.Conf.Version))

	zerolog.Debug().Interface("conf", setting.Conf).Go()
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

	if setting.Conf.Echo.CrosEnable {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: setting.Conf.Echo.CrosAllowOrigins,
			AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		}))
	}

	if setting.Conf.Echo.GzipEnable {
		e.Use(middleware.Gzip())
	}

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

	zerolog.Debug().Str("ver", setting.Conf.Version).Go()

	// Start server
	go func() {
		if err := e.Start(setting.Conf.Echo.Listen); err != nil {
			zerolog.Error().Err(err).Msg("echo start err")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		zerolog.Error().Err(err).Msg("echo Shutdown err")
	}
}
