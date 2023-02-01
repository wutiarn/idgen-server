package main

import (
	"github.com/ilyakaznacheev/cleanenv"
	"idgen-server/idgen"
	"os"
)

//goland:noinspection GoVetStructTag
type AppConfig struct {
	IdGen idgen.Config `yaml:"idGen"`
}

func GetConfig() *AppConfig {
	config := AppConfig{}
	configFile, configExists := os.LookupEnv("APP_CONFIG")
	var err error
	if configExists {
		err = cleanenv.ReadConfig(configFile, &config)
	} else {
		err = cleanenv.ReadEnv(&config)
	}

	if err != nil {
		panic(err)
	}
	return &config
}
