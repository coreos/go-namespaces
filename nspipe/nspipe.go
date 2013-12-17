package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	nameNet "github.com/coreos/go-namespaces/net"
)

var target *int = flag.Int("t", 0, "target pid")
var targetAddr *string = flag.String("l", "localhost:23", "local address")
var remoteAddr *string = flag.String("r", "towel.blinkenlights.nl:23", "remote address")

func proxyConn(conn *net.Conn) {
	rConn, err := net.Dial("tcp", *remoteAddr)
	if err != nil {
		panic(err)
	}

	go io.Copy(rConn, *conn)
	go io.Copy(*conn, rConn)
}

func main() {
	flag.Parse()

	if *target == 0 {
		fmt.Fprintln(os.Stderr, "error: a target pid is required")
		flag.PrintDefaults()
		return
	}

	fmt.Printf("PROXY: targetPid:%d targetAddr:%v remoteAddr:%v\n",
		*target, *targetAddr, *remoteAddr)

	listener, err := nameNet.ListenNamespace(uintptr(*target), "tcp", *targetAddr)
	if err != nil {
		panic(err)
	}

	for {
		// Wait for a connection.
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go proxyConn(&conn)
	}
}
