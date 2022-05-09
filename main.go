package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/averageNetAdmin/andproxy/cmd/config"
)

func main() {
	handlers, logDir, err := config.ReadAndCreate("/etc/andproxy/config.yml")
	if err != nil {
		log.Fatalln(err)
	}
	for _, handler := range handlers {
		handler.Handle()
	}
	err = os.MkdirAll(logDir, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	logFile, err := os.OpenFile(fmt.Sprintf("%s/andproxy.log", logDir), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	logger := log.New(logFile, "andproxy ", log.LstdFlags)
	logger.SetFlags(log.LstdFlags)

	listen, err := net.Listen("unix", "/run/andproxy.sock")
	if err != nil {
		logger.Fatalln(err)
	}
	defer listen.Close()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	end := false
	go func() {
		sig := <-signals
		switch sig {
		case os.Interrupt:
			for _, handler := range handlers {
				handler.SaveState()
				handler.Stop()
			}
			end = true
			err = listen.Close()
			if err != nil {
				logger.Fatal(err)
			}
		case os.Kill:
			for _, handler := range handlers {
				handler.SaveState()
				handler.Stop()
			}
			end = true
			err = listen.Close()
			if err != nil {
				logger.Fatalln(err)
			}
		default:
		}
	}()

	for {
		conn, err := listen.Accept()
		if end {
			break
		}
		if err != nil {
			logger.Println(err)
		}
		command := make([]byte, 100)
		n, err := conn.Read(command)
		if err != nil {
			logger.Println(err)
		}
		switch string(command[:n]) {
		case "update configuration":
			c, err := config.Read("/etc/andproxy/config.yml")
			if err != nil {
				conn.Write([]byte(err.Error()))
			}
			for name, config := range c {
				handlers[name].UpdateConfig(config)
			}
		case "get current state":
			data, err := json.Marshal(handlers)
			if err != nil {
				logger.Println(err)
			}
			_, err = conn.Write(data)
			if err != nil {
				logger.Println(err)
			}
		default:

		}
		err = conn.Close()
		if err != nil {
			logger.Println(err)
		}
	}
}
