package main

import (
	"fmt"
	"net"
	"strings"
)

var current = 0
var backends = []string{"localhost:8001", "localhost:8002"}

func handle(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 1024)

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error in reading bytes", err)
			return
		}

		message := strings.TrimSpace(string(buff[:n]))

		fmt.Printf("Routing %s to %s \n", message, backends[current])

		dbConn, err := net.Dial("tcp", backends[current])
		if err != nil {
			fmt.Printf("Error in connecting %s \n", backends[current])
			return
		}

		_, err = dbConn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error in sending packet to the db server")
		}

		readBuff := make([]byte, 1024)
		n, err = dbConn.Read(readBuff)
		if err != nil {
			fmt.Println("Error in reading from the db")
			return
		}

		conn.Write(readBuff[:n])

		current++
		if current >= len(backends) {
			current = 0
		}

	}

}

func main() {
	port := ":9000"

	listener, err := net.Listen("tcp", port)

	if err != nil {
		fmt.Println("Error in starting the server.")
		return
	}

	defer listener.Close()

	fmt.Printf("Listening on port %s...\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection")
			return
		}

		go handle(conn)
	}

}
