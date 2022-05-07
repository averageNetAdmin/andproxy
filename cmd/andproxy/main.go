package main

import (
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/andproxy.sock")
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte("update configuration"))
	conn.Close()
}
