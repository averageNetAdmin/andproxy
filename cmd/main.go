package main

import (
	"log"
	"time"

	"github.com/averageNetAdmin/andproxy/source/cnfrd"
)

func main() {
	handlers, err := cnfrd.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	for _, handler := range handlers {
		go handler.Handle()
	}
	for {
		time.Sleep(60 * time.Second)
	}
}
