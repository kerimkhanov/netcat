package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var (
	listen  = flag.Bool("l", false, "Listen")
	host    = flag.String("h", "localhost", "Host")
	port    = flag.Int("p", 8989, "Port")
	channel = make(chan string, 1)
	allUser = make(map[net.Conn]string)
	lastMsg = ""
	chat    = ""
)

func main() {
	flag.Parse()
	if *listen {
		startServer()
		return
	}
	if len(flag.Args()) < 2 {
		fmt.Println("Hostname and port required")
		return
	}
	serverHost := flag.Arg(0)
	serverPort := flag.Arg(1)
	startClient(fmt.Sprintf("%s:%s", serverHost, serverPort))
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
func startServer() {
	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
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
				item.Write([]byte(chat))
			}
		case <-dead:
			return
		}
	}
}

func startClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Can't connect to server: %s\n", err)
		return
	}
	fmt.Print("[ENTER YOUR NAME]:")
	reader := bufio.NewReader(os.Stdin)
	input, _, err := reader.ReadLine()
	if err != nil {
		fmt.Println(err)
	}
	// content, err := os.ReadFile(fileName)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(content))

	conn.Write(input)
	for {
		go readMessage(conn, string(input))
		writeMessage(conn, string(input))
	}
}

func readMessage(conn net.Conn, nick string) {

	for {
		input := make([]byte, 1024)
		n, err := conn.Read(input)
		if err != nil {
			break
		}

		if lastMsg != string(input[:n]) {
			if strings.TrimSpace(lastMsg) != "\n" {
				fmt.Printf("%s\n", string(input[:n]))
			}
		}
	}
}

func writeMessage(conn net.Conn, nick string) {
	reader := bufio.NewReader(os.Stdin)
	input, _, err := reader.ReadLine()

	if err != nil {
		fmt.Println(err)
	}
	currentTime := time.Now().Format("02-01-2006 15:04:05")
	lastMsg = fmt.Sprintf("[%s][%s]:%s\n", currentTime, nick, input)
	result := strings.TrimSpace(fmt.Sprintf("%s", input))
	fmt.Printf("result:-->%s<", result)
	if result != "" {
		// fmt.Printf("%#v", result)
		// fmt.Println("Hello -----")
		conn.Write([]byte(lastMsg))
	}
}
