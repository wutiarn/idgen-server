package main

import (
	"github.com/ilyakaznacheev/cleanenv"
	"idgen-server/idgen"
)

//goland:noinspection GoVetStructTag
type AppConfig struct {
	IdGen idgen.Config
}

func GetConfig() *AppConfig {
	config := AppConfig{}
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		panic(err)
	}
	return &config
}
