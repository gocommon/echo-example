package setting

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

// Config Config
type Config struct {
	Version string
	Echo    EchoService
}

// EchoService EchoService
type EchoService struct {
	Debug      bool
	HideBanner bool // 是否隐藏echo banner日志输出
}

// Conf Conf配置内容
var Conf = newConfig()

func newConfig() *Config {
	// 配置默认值，写在这！
	return &Config{Echo: EchoService{
		Debug:      true,
		HideBanner: true,
	}}
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
