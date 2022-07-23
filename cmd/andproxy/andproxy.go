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
	h, err := handler.NewHandler("/etc/andproxy/handlers/http_80")
	if err != nil {
		fmt.Println(err)
	}

	h.Listen()
	listen, err := net.Listen("unix", "/run/andproxy.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	endconn := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	go func() {
		for {
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
	memstats := &runtime.MemStats{}
	go func() {
		for {
			runtime.GC()
			runtime.ReadMemStats(memstats)
			fmt.Println(memstats.Alloc)
			time.Sleep(time.Second * 10)
		}

	}()
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
	for {

	}
	/*wg := new(sync.WaitGroup)
	start := time.Now()
	srv1 := 0
	srv2 := 0
	srv3 := 0
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {

			for ii := 0; ii < 100; ii++ {
				r, err := http.Get("http://192.168.31.222:80")
				if err != nil && err.Error() != "EOF" {
					fmt.Println(err)
				}
				info := make([]byte, 1500)
				_, err = r.Body.Read(info)
				if err != nil && err.Error() != "EOF" {
					fmt.Println(err)
				}
				if strings.Contains(string(info), "SERVER1") {
					srv1++
				} else if strings.Contains(string(info), "SERVER2") {
					srv2++
				} else if strings.Contains(string(info), "SERVER3") {
					srv3++
				}
				if ii%100 == 0 {
					fmt.Println(ii)
					time.Sleep(10 * time.Millisecond)
				}
				r.Body.Close()

			}
			wg.Done()
		}()

	}

	wg.Wait()
	dur := time.Since(start)
	fmt.Println(dur)
	fmt.Println(srv1, srv2, srv3)
	os.Remove("/run/andproxy.sock")
	os.Exit(1)*/
}
