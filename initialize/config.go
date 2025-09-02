package initialize

import (
	"github.com/spf13/viper"
	"oneclick-metrics-go/global"
)

// 初始化配置类
func InitConfig() {
	configFileName := "config.yaml"
	v := viper.New()
	v.SetConfigFile(configFileName)

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(global.ServerConfig); err != nil {
		panic(err)
	}

}
