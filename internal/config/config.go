package config

import (
	"bytes"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	DiscoveryPeriod   int    `mapstructure:"period"`
	WaitingTime       int    `mapstructure:"waiting"`
	Type              int    `mapstructure:"type"`
	ReliableUDPServer string `mapstructure:"addr"`
}

func Read() Config {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.SetConfigType("yml")

	if err := viper.ReadConfig(bytes.NewBufferString(Default)); err != nil {
		log.Fatalf("err: %s", err)
	}

	viper.SetConfigName("config")

	if err := viper.MergeInConfig(); err != nil {
		log.Print("No config file found")
	}

	viper.SetEnvPrefix("p2p")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("err: %s", err)
	}

	return cfg
}
