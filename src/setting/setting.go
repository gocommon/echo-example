package setting

import (
	"io/ioutil"
	"os"

	"modules/zerolog"

	"github.com/BurntSushi/toml"
)

// Config Config
type Config struct {
	Version string
	Echo    EchoService

	ZeroLogs map[string]map[string]zerolog.Option
}

// EchoService EchoService
type EchoService struct {
	Debug      bool
	HideBanner bool // 是否隐藏echo banner日志输出

	Listen string

	AccessLog         bool // 是否显示访问日志
	AccessLogFile     bool
	AccessLogFilePath string

	CrosEnable       bool
	CrosAllowOrigins []string

	GzipEnable bool
}

// Conf Conf配置内容
var Conf = newConfig()

func newConfig() *Config {
	// 配置默认值，写在这！
	return &Config{
		Echo: EchoService{
			// Debug:      true,
			HideBanner: true,
			Listen:     ":8899",
			AccessLog:  true,
			GzipEnable: true,
		},
		ZeroLogs: map[string]map[string]zerolog.Option{
			"default": {
				"console": {
					Enable: true,
					Mode:   "console",
					Level:  "debug",
				},
			},
		},
	}
}

// InitConf InitConf 初始化配置
func InitConf(confPath string) (err error) {

	contents, err := replaceEnvsFile(confPath)
	if err != nil {
		return err
	}

	if _, err = toml.Decode(contents, &Conf); err != nil {
		return err
	}

	return nil
}

func replaceEnvsFile(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return os.ExpandEnv(string(contents)), nil
}
