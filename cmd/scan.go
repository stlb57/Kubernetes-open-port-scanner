package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Target struct {
	PodName string
	IP      string
	Port    int
}

type ScanResult struct {
	PodName   string
	IP        string
	Port      int
	Timestamp string
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan K8s Pods for open ports",
	Long:  `Discovers Pod IPs in the cluster and scans them using a concurrent worker pool.`,
	Run: func(cmd *cobra.Command, args []string) {
		startScanner()
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func startScanner() {
	userHomeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(userHomeDir, ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	clientset, _ := kubernetes.NewForConfig(config)

	pods, _ := clientset.CoreV1().Pods(Namespace).List(context.Background(), metav1.ListOptions{})

	jobs := make(chan Target, 100)
	results := make(chan ScanResult, 100)
	var wg sync.WaitGroup

	for w := 1; w <= 50; w++ {
		wg.Add(1)
		go port_checker(&wg, jobs, results)
	}

	go func() {
		commonPorts := []int{80, 443, 8080}
		for _, pod := range pods.Items {
			if pod.Status.PodIP != "" {
				for _, p := range commonPorts {
					jobs <- Target{PodName: pod.Name, IP: pod.Status.PodIP, Port: p}
				}
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Printf("🔍 Scanning namespace: %s...\n", Namespace)
	for res := range results {
		fmt.Println(res)
	}
}

func port_checker(wg *sync.WaitGroup, jobs <-chan Target, results chan<- ScanResult) {
	defer wg.Done()
	for target := range jobs {
		address := fmt.Sprintf("%s:%d", target.IP, target.Port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			res := ScanResult{
				PodName:   target.PodName,
				IP:        target.IP,
				Port:      target.Port,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}

			results <- res
			conn.Close()
		}
	}
}
