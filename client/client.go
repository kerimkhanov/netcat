package client

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

var (
	channel = make(chan string, 1)
	allUser = make(map[net.Conn]string)
	lastMsg = ""
	chat    = ""
)

func startClient(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Can't connect to server: %s\n", err)
		return
	}
	body, err := ioutil.ReadFile("/logo/linuxlogo.txt")
	if err != nil {
		fmt.Errorf("Linux logo file not correct read")
	}
	fmt.Println(string(body))
	fmt.Print("[ENTER YOUR NAME]:")
	reader := bufio.NewReader(os.Stdin)
	input, _, err := reader.ReadLine()
	if err != nil {
		fmt.Println(err)
	}
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
			fmt.Printf("%s", string(input[:n]))
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
	lastMsg = fmt.Sprintf("\r[%s]:[%s]:%s\n", currentTime, nick, input)
	result := strings.TrimSpace(fmt.Sprintf("%s", input))
	if result != "" {
		conn.Write([]byte(lastMsg))
	}
}
