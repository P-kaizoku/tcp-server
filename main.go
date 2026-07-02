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
	walFile   *os.File
)

func load_data() {
	data, err := os.ReadFile("data.log")
	if err != nil {
		return
	}

	logs := strings.Split(string(data), "\n")

	for _, log := range logs {
		if log == "" {
			continue
		}
		parts := strings.Fields(log)

		if len(parts) >= 3 && parts[0] == "SET" {
			store[parts[1]] = strings.Join(parts[2:], " ")
		}
	}

}

func saveFile(key, val string) {
	walFile.WriteString(fmt.Sprintf("SET %s %s\n", key, val))
	walFile.Sync()
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
			parts := strings.Fields(strings.TrimSpace(string(buff[:n])))

			if len(parts) >= 3 && parts[0] == "SET" {
				key := parts[1]
				val := strings.Join(parts[2:], " ")

				mu.Lock()
				store[key] = val
				mu.Unlock()

				saveFile(key, val)
				conn.Write([]byte("Key set successfully\r\n"))
			} else {
				conn.Write([]byte("Error: missing key or val\r\n"))
			}

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
	var err error
	walFile, err = os.OpenFile("data.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer walFile.Close()

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
	load_data()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection", err)
			continue
		}

		go handle(conn)

	}
}
