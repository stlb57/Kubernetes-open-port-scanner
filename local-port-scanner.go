package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Target struct {
	IP   string
	Port int
}

func port_checker(wg *sync.WaitGroup, jobs <-chan Target, results chan<- string) {
	defer wg.Done()
	for target := range jobs {
		address := fmt.Sprintf("%s:%d", target.IP, target.Port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			if target.Port == 80 {
				fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
				status, _ := bufio.NewReader(conn).ReadString('\n')
				results <- fmt.Sprintf("Port %d open on %s (banner: %s)", target.Port, target.IP, status)
			} else {
				results <- fmt.Sprintf("Port %d is open on %s", target.Port, target.IP)
			}
			conn.Close()
		}
	}
}

func main() {
	userHomeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(userHomeDir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		fmt.Printf("Error building config: %s\n", err.Error())
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	pods, _ := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})

	jobs := make(chan Target, 100)
	results := make(chan string, 100)
	var wg sync.WaitGroup

	for w := 1; w <= 100; w++ {
		wg.Add(1)
		go port_checker(&wg, jobs, results)
	}

	go func() {
		for _, pod := range pods.Items {
			if pod.Status.PodIP != "" {
				jobs <- Target{IP: pod.Status.PodIP, Port: 80}
			}
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
