package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"

	// "os"
	"time"
)

func port_checker(wg *sync.WaitGroup, jobs <-chan int, results chan<- string) {
	defer wg.Done()
	for port := range jobs {
		address := fmt.Sprintf("localhost:%d", port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			if port == 80 {
				fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
				status, _ := bufio.NewReader(conn).ReadString('\n')
				results <- fmt.Sprintf("Port %d open (banner: %s)", port, status)
			} else {
				results <- fmt.Sprintf("Port %d is open", port)
			}
			conn.Close()
		}
	}

}
func main() {
	jobs := make(chan int, 100)
	results := make(chan string, 100)
	var wg sync.WaitGroup
	// for port := 1; port <= 1024; port++ {
	// 	wg.Add(1)
	// 	go port_checker(port, &wg)
	// }

	for w := 1; w <= 100; w++ {
		wg.Add(1)
		go port_checker(&wg, jobs, results)
	}
	go func() {
		for port := 1; port <= 1000; port++ {
			jobs <- port
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		fmt.Println(res)
	}

}
