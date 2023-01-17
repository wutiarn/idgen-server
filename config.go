package main

import (
	"github.com/ilyakaznacheev/cleanenv"
)

//goland:noinspection GoVetStructTag
type AppConfig struct {
	NodeId uint8 `env:"NODE_ID" env-required`
}

func GetConfig() *AppConfig {
	config := AppConfig{
		NodeId: 1,
	}
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		panic(err)
	}
	return &config
}
