package config

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

var Glogger *logrus.Logger

func StartLogging() {
	f, err := os.OpenFile(AndConfig.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	Glogger = logrus.New()
	Glogger.SetFormatter(&logrus.JSONFormatter{})
	Glogger.SetOutput(f)
	Glogger.SetLevel(logrus.InfoLevel)
}
