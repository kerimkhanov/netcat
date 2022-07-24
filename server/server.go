package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	channel = make(chan string, 1)
	allUser = make(map[net.Conn]string)
	host    = flag.String("h", "localhost", "Host")
	port    = flag.Int("p", 8989, "Port")
)

func StartServer() {

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
		if len(allUser) > 10 {
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
	for _, users := range allUser {
		if users == string(input[:n]) {
			conn.Write([]byte("Your nickname is busy"))
			conn.Close()
			return
		}
	}
	read, err := os.ReadFile(f.Name())
	if err != nil {
		fmt.Println(err)
	}
	conn.Write([]byte(strings.TrimRight(string(read), "\n")))
	allUser[conn] = string(input[:n])
	channel <- fmt.Sprintf("\n%s join\n", allUser[conn])
	dead := make(chan bool, 1)
	go toChannel(conn, dead)
	go fromChannel(conn, dead, f)
}

func toChannel(conn net.Conn, dead chan bool) {
	for {
		input := make([]byte, 1024)
		n, err := conn.Read(input)
		if err != nil {
			channel <- fmt.Sprintf("\n%s left\n", allUser[conn])
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
			f.WriteString(strings.TrimLeft(msg, "\n"))
			for item := range allUser {
				item.Write([]byte(msg))
				time.Sleep(time.Second / 1000)
				item.Write([]byte(fmt.Sprintf("[%s]:[%s]:", time.Now().Format("02-01-2006 15:04:05"), allUser[item])))
			}
		case <-dead:
			return
		}
	}
}
