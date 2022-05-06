package main

import (
	"log"
	"sync"

	"github.com/averageNetAdmin/andproxy/cmd/cnfrd"
)

func main() {
	var wg sync.WaitGroup
	handlers, err := cnfrd.ReadConfig("/etc/andproxy/config.yml")
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(len(handlers))
	for _, handler := range handlers {
		handler.Handle()
	}
	wg.Wait()
}
