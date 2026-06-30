package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	store     = make(map[string]string)
	taskQueue = make(chan string, 100)
	mu        sync.RWMutex
)

func 

func saveFile(key, val string) {
	file, err := os.OpenFile("data.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("SET %s %s", key, val))
}

func handle(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 1024)

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error in reading bytes", err)
			return
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
				fmt.Println("key or val is missing")
				conn.Write([]byte("key or val is missing"))
				continue
			}
			mu.Lock()
			store[key] = val
			mu.Unlock()
			saveFile(key, val)
			conn.Write([]byte("Key set successfully\r\n"))

		case "GET":
			key := message[1]
			mu.RLock()
			val, ok := store[key]
			mu.RUnlock()
			if !ok {
				conn.Write([]byte("Error: key not found"))
				continue
			}
			buff := []byte(val + "\r\n")
			conn.Write(buff)
		case "TASK":
			conn.Write([]byte("Task added into the queue\r\n"))
			msg := message[1]
			if msg == "" {
				fmt.Println("Please enter your task")
				continue
			}
			taskQueue <- msg
		}
	}

}

func main() {

	fmt.Printf("The workers are ready and waiting...")
	for i := 1; i <= 3; i++ {
		go func(workerID int) {

			for task := range taskQueue {
				fmt.Printf("the %s is processing...", task)
				time.Sleep(2 * time.Second)
				fmt.Printf("Worder %d has finish %s", workerID, task)
			}
		}(i)
	}

	

	port := ":8080"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	listener, err := net.Listen("tcp", port)
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
