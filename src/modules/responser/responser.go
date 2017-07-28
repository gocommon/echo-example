package responser

import (
	"strings"

	"github.com/labstack/echo"
)

var (
	// DefaultFormatParam 默认query返回格式参数
	DefaultFormatParam = "_resfmt"

	// DefaultJsonpParam 默认jsonp 返回函数名
	DefaultJsonpParam = "callback"

	// DefaultStringResponseField 默认字符串返回值字段
	DefaultStringResponseField = "message"
)

// R 统计一返回值方法
func R(c echo.Context, code int, i interface{}) error {

	req := c.Request()

	resFormat := strings.ToLower(c.QueryParam(DefaultFormatParam))

	ctype := req.Header.Get(echo.HeaderContentType)
	switch {
	case resFormat == "json":

		return c.JSON(code, i)

	case resFormat == "jsonp":

		return c.JSONP(code, c.QueryParam(DefaultJsonpParam), i)

	case resFormat == "xml":

		return c.XML(code, i)

	case strings.HasPrefix(ctype, echo.MIMEApplicationJSON):

		return c.JSON(code, i)

	case strings.HasPrefix(ctype, echo.MIMEApplicationXML), strings.HasPrefix(ctype, echo.MIMETextXML):

		return c.XML(code, i)

	}

	// 字符串处理
	if _, ok := i.(string); ok {
		return c.JSON(code, echo.Map{DefaultStringResponseField: i})
	}
	// 默认返回json
	return c.JSON(code, i)
}
