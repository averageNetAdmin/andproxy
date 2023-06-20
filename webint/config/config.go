package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type AndConifg struct {
	LogPath     string `yaml:"logPath"`
	LogLevel    string `yaml:"logLevel"`
	Port        int32  `yaml:"port"`
	AndProxyUrl string `yaml:"andProxyUrl"`
}

var AndConfig = &AndConifg{}

func ReadFile(path string) {

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, AndConfig)
	if err != nil {
		log.Fatal(err)
	}
}
