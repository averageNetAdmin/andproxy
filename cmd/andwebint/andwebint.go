package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := net.Dial("unix", "/run/andproxy.sock")
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte("get current state"))
	data := make([]byte, 5000)
	n, err := conn.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, string(data[:n]))
}

func main() {

	hndlr := new(handler)
	http.ListenAndServe(":8000", hndlr)

}
