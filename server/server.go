package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	listen  = flag.Bool("l", false, "Listen")
	host    = flag.String("h", "localhost", "Host")
	port    = flag.Int("p", 8989, "Port")
	channel = make(chan string, 1)
	allUser = make(map[net.Conn]string)
)

func startServer() {
	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	fileName := fmt.Sprintf("%s %d.txt", time.Now().Format("02-01-2006 15:04:05"), *port)
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Printf("Listening for connections on %s", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if len(allUser) > 2 {
			conn.Write([]byte("Sorry, your connection is close. More than 10 users"))
			conn.Close()
		}
		if err != nil {
			log.Printf("Error accepting connection from client: %s", err)
		}
		go processClient(conn, f)

	}
}

func processClient(conn net.Conn, f *os.File) {
	input := make([]byte, 1024)
	n, _ := conn.Read(input)
	read, err := os.ReadFile(f.Name())
	for _, w := range read {
		fmt.Printf("%v", w)
	}
	if err != nil {
		fmt.Println(err)
	}
	conn.Write(read)
	allUser[conn] = string(input[:n])
	channel <- fmt.Sprintf("%s join\n", allUser[conn])
	dead := make(chan bool, 1)
	go toChannel(conn, dead)
	go fromChannel(conn, dead, f)
}

func toChannel(conn net.Conn, dead chan bool) {
	for {
		input := make([]byte, 1024)
		n, err := conn.Read(input)
		if err != nil {
			channel <- fmt.Sprintf("%s left\n", allUser[conn])
			dead <- true
			delete(allUser, conn)
			return
		}
		channel <- fmt.Sprintf("%s", string(input[:n]))
	}
}
func fromChannel(conn net.Conn, dead chan bool, f *os.File) {
	for {
		select {
		case msg := <-channel:
			f.WriteString(msg)
			for item := range allUser {
				item.Write([]byte(msg))
				// item.Write(strings.Trim([]byte(fmt.Sprintf("[%s:%s]:", time.Now().Format("02-01-2006 15:04:05"), allUser[item])))
				item.Write([]byte(fmt.Sprintf("\r[%s:%s]:", time.Now().Format("02-01-2006 15:04:05"), allUser[item])))
			}
		case <-dead:
			return
		}
	}
}
