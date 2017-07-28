package zerolog

import (
	"strings"

	"io"

	"fmt"

	"github.com/gocommon/rotatefile"
	"github.com/gocommon/zerolog"
	"github.com/gocommon/zerolog/op"
)

// Option Option
type Option struct {
	Enable bool
	Mode   string
	Level  string

	// model file
	FileName string // 文件名
	// LogRotate    bool   // 分割文件
	// MaxLines     int    // 最大行数
	// MaxSizeShift int    // 最大文件大小 1 << MaxSizeShift
	// DailyRotate  bool   // 每天分割文件
	// MaxDays      int    // 分割文件保留天数

	// model smtp
	User      string
	Passwd    string
	Host      string
	Receivers []string // default "[]"
	Subject   string
}

// Levels Levels
var Levels = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

// LevelByString LevelByString
func LevelByString(str string) zerolog.Level {
	str = strings.ToLower(str)

	if l, has := Levels[str]; has {
		return l
	}

	return zerolog.DebugLevel
}

// Loggers Loggers
var Loggers = map[string]zerolog.Logger{}

// WithContext WithContext
type WithContext func(zerolog.Context) zerolog.Context

// Timestamp Timestamp
func Timestamp() WithContext {
	return func(c zerolog.Context) zerolog.Context {
		return c.Timestamp()
	}
}

// Version Version
func Version(ver string) WithContext {
	return func(c zerolog.Context) zerolog.Context {
		return c.Str("version", ver)
	}
}

// InitLog InitLog
func InitLog(confs map[string]map[string]Option, withs ...WithContext) {

	for name := range confs {

		writers := make([]io.Writer, 0, len(confs[name]))

		for k := range confs[name] {
			opt := confs[name][k]
			if !opt.Enable {
				continue
			}

			switch strings.ToLower(opt.Mode) {
			case "console":

				writers = append(writers, op.NewConsole(LevelByString(opt.Level)))

			case "file":
				fd, err := rotatefile.NewWriter(rotatefile.Options{Filename: opt.FileName})
				if err != nil {
					panic(err)
				}

				writers = append(writers, op.NewFileWriter(fd, LevelByString(opt.Level)))

			case "smtp":
				writers = append(writers, op.NewSmtpWriter(opt.User, opt.Passwd, opt.Host, opt.Subject, opt.Receivers, LevelByString(opt.Level)))
			}
		}

		c := zerolog.New(zerolog.MultiLevelWriter(writers...)).With()

		for i := range withs {
			c = withs[i](c)
		}

		Loggers[name] = c.Logger()
	}

}

// Get Get
func Get(name ...string) zerolog.Logger {
	cname := "default"
	if len(name) > 0 {
		cname = name[0]
	}
	if logger, has := Loggers[cname]; has {
		return logger
	}

	panic(fmt.Sprintf("log.Get miss name %s, forget to Initlog ?", name))
}

// With With
func With() zerolog.Context {
	return Get().With()
}

// Debug Debug
func Debug() *zerolog.Event {
	return Get().Debug()
}

// Info Info
func Info() *zerolog.Event {
	return Get().Info()
}

// Warn Warn
func Warn() *zerolog.Event {
	return Get().Warn()
}

// Error Error
func Error() *zerolog.Event {
	return Get().Error()
}

// Fatal Fatal
func Fatal() *zerolog.Event {
	return Get().Fatal()
}

// Panic Panic
func Panic() *zerolog.Event {
	return Get().Panic()
}
