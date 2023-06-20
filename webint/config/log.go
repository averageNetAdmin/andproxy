package config

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func StartLogging() {
	f, err := os.OpenFile(AndConfig.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetOutput(f)
	Logger.SetLevel(logrus.InfoLevel)
}
