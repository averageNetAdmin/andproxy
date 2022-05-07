package main

import (
	"log"
	"net"

	"github.com/averageNetAdmin/andproxy/cmd/config"
)

func main() {
	handlers, err := config.ReadAndCreate("/etc/andproxy/config.yml")
	if err != nil {
		log.Fatal(err)
	}
	for _, handler := range handlers {
		handler.Handle()
	}
	listen, err := net.Listen("unix", "/tmp/andproxy.sock")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		command := make([]byte, 100)
		n, err := conn.Read(command)
		if err != nil {
			log.Fatal(err)
		}
		switch string(command[:n]) {
		case "update configuration":
			c, err := config.Read("/etc/andproxy/config.yml")
			if err != nil {
				log.Fatal(err)
			}
			for name, config := range c {
				handlers[name].UpdateConfig(config)
			}
		default:
			continue
		}
	}
}
