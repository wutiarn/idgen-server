package main

import (
	"github.com/spf13/viper"
	"strings"
)

type AppConfig struct {
	NodeId int8
}

func GetConfig() *AppConfig {
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("IDGEN")
	viper.AutomaticEnv()
	config := AppConfig{
		NodeId: 1,
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	return &config
}
