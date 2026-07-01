package main

import (
	"fmt"
	"net"
	"strings"
)

var (
	current  = 0
	backends = []string{"localhost:8001", "localhost:8002"}
	pool     = []chan net.Conn{make(chan net.Conn, 10), make(chan net.Conn, 10)}
)

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

		dbConn := <-pool[current]

		_, err = dbConn.Write([]byte(message + "\r\n"))
		if err != nil {
			fmt.Println("Error in sending packet to the db server")
			pool[current] <- dbConn
			return
		}

		readBuff := make([]byte, 1024)
		n, err = dbConn.Read(readBuff)
		if err != nil {
			fmt.Println("Error in reading from the db")
			pool[current] <- dbConn
			return
		}

		pool[current] <- dbConn

		conn.Write(readBuff[:n])

		current++
		if current >= len(backends) {
			current = 0
		}

	}

}

func createPool(connectionString string, connectionPool chan net.Conn, poolNo int) {

	for i := 0; i < cap(connectionPool); i++ {

		dbConn, err := net.Dial("tcp", connectionString)
		if err != nil {
			fmt.Printf("Error in creating %s connection %d pool\n", connectionString, poolNo)
			return
		}

		connectionPool <- dbConn
		fmt.Printf("connection %d successful in pool %d\n", i, poolNo)
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

	for i := range 2 {
		createPool(backends[i], pool[i], i)
	}

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
