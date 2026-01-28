package config

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Wal                 WALConfig      `mapstructure:"wal" yaml:"wal"`
	CoinPairGroups      map[string]int `mapstructure:"coin_pair_groups" yaml:"coin_pair_groups"`
	CoinPairPathMapping map[int]string `mapstructure:"coin_pair_path_mapping" yaml:"coin_pair_path_mapping"`
}

type WALConfig struct {
	FullLogsPrePath       string `mapstructure:"full_logs_prepath" yaml:"full_logs_prepath"`
	IncrementalLogPrePath string `mapstructure:"incremental_log_prepath" yaml:"incremental_log_prepath"`
}

var GlobalConf = Config{}

func init() {
	env := os.Getenv("env")
	if env == "" {
		panic("env must be specified")
	}
	configFilePath := fmt.Sprintf("./config/config.%s.yaml", env)

	v := viper.New()
	v.SetConfigFile(configFilePath)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(&GlobalConf); err != nil {
		panic(err)
	}

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		_ = v.ReadInConfig()
		_ = v.Unmarshal(&GlobalConf)
	})

}
