package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/averageNetAdmin/andproxy/internal/handler"
)

func main() {
	// Create handler from file and run it
	// TODO: add code for reading all handlers from directory
	h, err := handler.NewHandler("/etc/andproxy/handlers/http_80")
	if err != nil {
		fmt.Println(err)
	}
	h.Listen()
	
	// Open socket to excange data with other programs
	// It`s for web interface 
	listen, err := net.Listen("unix", "/run/andproxy.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()
	
	// handle system signals
	endconn := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	go func() {
		for {	
			// For correct handling Ctrl+C and other signals that end program
			// If not remove andproxy.sock program will not start until file exist
			// TODO: add different behavior for different signals
			sig := <-signals
			switch sig {
			case syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT:
				endconn <- struct{}{}
				os.Remove("/run/andproxy.sock")
				os.Exit(0)
			case os.Interrupt:
				endconn <- struct{}{}
				os.Remove("/run/andproxy.sock")
				os.Exit(0)
			case os.Kill:
				endconn <- struct{}{}
				os.Remove("/run/andproxy.sock")
				os.Exit(0)
			default:
			}
		}

	}()
	// mem test **will be deleted in prod**
	memstats := &runtime.MemStats{}
	go func() {
		for {
			runtime.GC()
			runtime.ReadMemStats(memstats)
			fmt.Println(memstats.Alloc)
			time.Sleep(time.Second * 10)
		}

	}()
	
	// handle socket requests
	// send current state of all handlers
	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				log.Println(err)
			}
			command := make([]byte, 100)
			n, err := conn.Read(command)
			if err != nil {
				log.Println(err)
			}
			switch string(command[:n]) {
			case "get current state":
				// only send current state handler object
				// all validation on outside
				data, err := json.Marshal(h)
				if err != nil {
					log.Println(err)
				}
				_, err = conn.Write(data)
				if err != nil {
					log.Println(err)
				}
			default:

			}
			err = conn.Close()
			if err != nil {
				log.Println(err)
			}
		}

	}()
	// without this program immediately end
	for {

	}
	
}
