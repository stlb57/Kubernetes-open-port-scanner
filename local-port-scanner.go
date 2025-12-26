package main

import (
	"bufio"
	"fmt"
	"net"

	// "os"
	"time"
)

func port_checker(port int, c chan string) {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err == nil {
		if port == 80 {
			fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
			status, err := bufio.NewReader(conn).ReadString('\n')
			fmt.Println(status, err)
		}
		c <- "Port open"
		conn.Close()
	}
}
func main() {
	c := make(chan string)
	for port := 1; port <= 1024; port++ {
		go port_checker(port, c)
	}
	for port := 1; port <= 1024; port++ {
		fmt.Println(<-c)
	}

}
