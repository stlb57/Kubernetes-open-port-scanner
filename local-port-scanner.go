package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"

	// "os"
	"time"
)

func port_checker(port int, wg *sync.WaitGroup) {
	defer wg.Done()
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err == nil {
		if port == 80 {
			fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
			status, err := bufio.NewReader(conn).ReadString('\n')
			fmt.Println(status, err)
		} else {
			fmt.Printf("Port %d is open", port)
		}
		conn.Close()
	}
}
func main() {
	var wg sync.WaitGroup
	for port := 1; port <= 1024; port++ {
		wg.Add(1)
		go port_checker(port, &wg)
	}
	wg.Wait()

}
