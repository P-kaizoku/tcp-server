package main

import (
	"fmt"
	"net"
	"strings"
)

func handle(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 1024)

	n, err := conn.Read(buff)
	if err != nil {
		fmt.Println("Error in reading bytes", err)
		return
	}

	message := strings.TrimSpace(string(buff[:n]))
	fmt.Println("Recieved:", message)

	if message == "PING" {
		conn.Write([]byte("PING\r\n"))
	} else {
		conn.Write([]byte("Invalid command\r\n"))
	}

}

func main() {
	listener, _ := net.Listen("localhost", ":8080")

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
