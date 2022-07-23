package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func testConn() {
	wg := new(sync.WaitGroup)
	start := time.Now()
	srv1 := 0
	srv2 := 0
	srv3 := 0
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {

			for ii := 0; ii < 1000; ii++ {
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
}

type H struct {
}

func (h *H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func listen() {
	h := new(H)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", "8000"),
		Handler: h,
	}

	err := server.ListenAndServe()
	if err != nil && err.Error() != "EOF" {
		fmt.Println(err)
	}
}

func main() {
	listen()
	for {

	}
}
