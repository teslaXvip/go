package concurrence

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// config 配置结构体
type config struct {
	Password      string
	ServerAddress string
}

var (
	cfg *config
	mu1 sync.Mutex
)

func GetConfig() *config {
	if cfg == nil {
		mu1.Lock()
		defer mu1.Unlock()
		if cfg == nil {
			vp := viper.New()
			vp.AddConfigPath("../../config") // 相对test file的路径
			vp.SetConfigName("mysql")
			vp.SetConfigType("yaml")

			fmt.Println("解析配置文件")
			if err := vp.ReadInConfig(); err != nil {
				fmt.Println(err)
				return nil
			} else {
				cfg = &config{
					Password:      vp.GetString("gift.pass"),
					ServerAddress: vp.GetString("gift.host"),
				}
			}
		}
	}
	return cfg
}
