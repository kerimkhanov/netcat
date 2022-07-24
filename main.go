package main

import (
	"flag"
	"fmt"
	"netcat/client"
	"netcat/server"
)

var (
	listen = flag.Bool("l", false, "Listen")
)

func main() {
	flag.Parse()
	if *listen {
		server.StartServer()
		return
	}
	if len(flag.Args()) < 2 {
		fmt.Println("Hostname and port required")
		return
	}
	serverHost := flag.Arg(0)
	serverPort := flag.Arg(1)
	client.StartClient(fmt.Sprintf("%s:%s", serverHost, serverPort))
}
