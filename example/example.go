package main

import (
	"net/http"
	"io"
	"os"
	"strconv"

	"github.com/philips/go-namespace/net"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello socket namespace world!\n")
}

func main() {
	args := os.Args

	out := http.NewServeMux()
	out.HandleFunc("/", HelloServer)
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        out,
	}
	go srv.ListenAndServe()


	pid, _ := strconv.Atoi(args[1])
	l, err := net.ListenNamespace(uintptr(pid))
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", HelloServer)
	go http.Serve(l, nil)

	out = http.NewServeMux()
	out.HandleFunc("/", HelloServer)
	srv = &http.Server{
		Addr:           ":8081",
		Handler:        out,
	}
	srv.ListenAndServe()
}
