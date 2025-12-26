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

// Target struct for the worker pool
type Target struct {
	IP   string
	Port int
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan K8s Pods for open ports",
	Long:  `Discovers Pod IPs in the cluster and scans them using a concurrent worker pool.`,
	Run: func(cmd *cobra.Command, args []string) {
		startScanner() // This calls your actual logic below
	},
}

func init() {
	rootCmd.AddCommand(scanCmd) // This connects 'scan' to your main 'k8s-scan' tool [cite: 1921]
}

func startScanner() {
	// 1. Setup K8s Client [cite: 1289, 1293]
	userHomeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(userHomeDir, ".kube", "config")
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	clientset, _ := kubernetes.NewForConfig(config)

	// 2. Fetch Pods using the Namespace variable from root.go [cite: 1314, 1318]
	pods, _ := clientset.CoreV1().Pods(Namespace).List(context.Background(), metav1.ListOptions{})

	// 3. Worker Pool Setup [cite: 106, 110, 552]
	jobs := make(chan Target, 100)
	results := make(chan string, 100)
	var wg sync.WaitGroup

	for w := 1; w <= 50; w++ { // Hire 50 workers [cite: 578]
		wg.Add(1)
		go port_checker(&wg, jobs, results)
	}

	// 4. Feeding the Machine [cite: 583, 584, 588]
	go func() {
		commonPorts := []int{80, 443, 8080}
		for _, pod := range pods.Items {
			if pod.Status.PodIP != "" {
				for _, p := range commonPorts {
					jobs <- Target{IP: pod.Status.PodIP, Port: p}
				}
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()      // Wait for workers [cite: 590, 605]
		close(results) // Close results bin [cite: 592]
	}()

	// 5. Output
	fmt.Printf("🔍 Scanning namespace: %s...\n", Namespace)
	for res := range results {
		fmt.Println(res)
	}
}

func port_checker(wg *sync.WaitGroup, jobs <-chan Target, results chan<- string) {
	defer wg.Done()
	for target := range jobs {
		address := fmt.Sprintf("%s:%d", target.IP, target.Port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			results <- fmt.Sprintf("✅ OPEN | %s:%d", target.IP, target.Port)
			conn.Close()
		}
	}
}
