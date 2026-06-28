package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func handle(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 1024)

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error in reading bytes", err)
		}

		message := strings.Fields(strings.TrimSpace(string(buff[:n])))
		fmt.Println("Recieved:", message)

		switch message[0] {
		case "PING":
			conn.Write([]byte("PONG\r\n"))

		case "SET":
			key := message[1]
			val := message[2]
			if key == "" || val == "" {
				fmt.Print("key or val is missing")
				conn.Write([]byte("key or val is missing"))
			}
			mu.Lock()
			store[key] = val
			mu.Unlock()
			conn.Write([]byte("Key set successfully\r\n"))

		case "GET":
			key := message[1]
			mu.RLock()
			val, ok := store[key]
			mu.RUnlock()
			if !ok {
				conn.Write([]byte("Error: key not found"))
			}
			buff := []byte(val + "\r\n")
			conn.Write(buff)
		}
	}

}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("error occured in starting the server", err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}

		go handle(conn)

	}
}
