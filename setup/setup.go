package setup

import (
	"flag"
	"github.com/spf13/viper"
	"gockets/src/services/logger"
	"gockets/src/services/tickerHelper"
)

func init() {
}

func viperDefaultValues() {
	viper.SetDefault("managementProtect", true)
	viper.SetDefault("allowedHostnames", []string{
		"127.0.0.1",
		"localhost",
	})
	viper.SetDefault("pingInterval", 10)
	viper.SetDefault("logLevel", 3)
	viper.SetDefault("port", 8844)
}

func Init() {
	flagPass()
	viperDefaultValues()
	initPackages()
	ll.Log.Debug("Setup complete")
}

func flagPass() {
	configPtr := flag.String("configPath", "config.yml", "Path to config file")
	flag.Parse()
	viper.AddConfigPath(*configPtr)
}

func initPackages() {
	tickerHelper.SetPingInterval(viper.GetInt("pingInterval"))
	ll.InitLogger(viper.GetInt("logLevel"))
}
