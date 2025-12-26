package main

import (
	"bufio"
	"fmt"
	"net"
	// "os"
	// "time"
)

func main() {
	for port := 1; port <= 20; port++ {
		address := fmt.Sprintf("localhost:%d", port)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Println("Port closed")
		} else {
			fmt.Println("Port open")
			fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
			status, err := bufio.NewReader(conn).ReadString('\n')
			fmt.Println(status, err)
			defer conn.Close()
		}
	}

}
