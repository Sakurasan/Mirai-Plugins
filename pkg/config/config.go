package config

import (
	"io"
	"os"

	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	AppTpl = `#默认配置
bot:
  loginmethod: qrcode # common,qrcode
  account: ""
  password: ""`
	PluginTpl = ``
	gconfname = "bot.yaml"
	logger    = utils.GetModuleLogger("plugin.config")
)

func init() {
	if !fileutil.IsExist("./plugins.yaml") && !fileutil.IsExist("./config/plugins.yaml") {
		f, _ := os.OpenFile("plugins.yaml", os.O_CREATE|os.O_RDWR, os.ModePerm)
		defer f.Close()
		f.WriteString("plugins:\n")
	}
}

type Config struct {
	*viper.Viper
}

type opt func(*Config)

func WithConfigName(name string) opt {
	gconfname = name
	return func(v *Config) {
		v.SetConfigName(name)
	}
}

func WithConfigPath(path string) opt {
	return func(v *Config) {
		v.AddConfigPath(path)
	}
}

// GlobalConfig 默认全局配置
var PluginConfig *Config

// Init 使用 ./plugins.yaml 初始化全局配置
func Init(opts ...opt) {
	PluginConfig = &Config{
		viper.New(),
	}
	PluginConfig.SetConfigName("plugins")
	PluginConfig.SetConfigType("yaml")
	PluginConfig.AddConfigPath(".")
	PluginConfig.AddConfigPath("./config")
	PluginConfig.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("PluginConfig file changed:", e.Name)
	})
	PluginConfig.WatchConfig()
	for _, o := range opts {
		o(PluginConfig)
	}

	err := PluginConfig.ReadInConfig()
	if err != nil {
		logger.WithField("config", "PluginConfig").WithError(err).Panicf("unable to read plugins config")
	}
}

func GlobalConfigInit(opts ...opt) {
	lgconfig := &Config{
		Viper: viper.New(),
	}

	lgconfig.SetConfigName(gconfname)
	lgconfig.SetConfigType("yaml")
	lgconfig.AddConfigPath(".")
	lgconfig.AddConfigPath("./config")
	lgconfig.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("GlobalConfig file changed:", e.Name)
	})
	lgconfig.WatchConfig()
	for _, o := range opts {
		o(lgconfig)
	}

	config.GlobalConfig = &config.Config{
		Viper: lgconfig.Viper,
	}

	if !fileutil.IsExist(gconfname) {
		f, _ := os.OpenFile(gconfname, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
		defer f.Close()
		io.WriteString(f, AppTpl)
	}

	err := config.GlobalConfig.ReadInConfig()
	if err != nil {
		logger.WithField("config", "GlobalConfig").WithError(err).Panicf("unable to read global config =>%s", gconfname)
	}
}
