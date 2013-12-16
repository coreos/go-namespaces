package main

import (
	"net/http"
	"io"
	"os"
	"strconv"

	"github.com/philips/go-namespace/net"
)

func HelloPid(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello pid namespace!\n")
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello original namespace!\n")
}

// This example application creates an http listening inside of the namespace
// of the process given in os.Args[0] on port 8080 and an http server in the
// original namespace each with different messages.
func main() {
	args := os.Args

	pid, _ := strconv.Atoi(args[1])
	l, err := net.ListenNamespace(uintptr(pid), "tcp", ":8080")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", HelloPid)
	go http.Serve(l, nil)

	out := http.NewServeMux()
	out.HandleFunc("/", HelloServer)
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        out,
	}
	srv.ListenAndServe()
}
