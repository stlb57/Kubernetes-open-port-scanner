package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	scansTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "k8s_port_scans_total",
		Help: "The total number of pods scanned for open ports",
	})

	openPortsFound = promauto.NewCounter(prometheus.CounterOpts{
		Name: "k8s_open_ports_total",
		Help: "The total number of open ports discovered",
	})

	scanDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "k8s_scan_duration_seconds",
		Help:    "Duration of port scans in seconds",
		Buckets: prometheus.DefBuckets,
	})

	k8s_active_workers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_active_workers",
		Help: "Gauge to track the current load on your worker pool.",
	})
)

type MockScanner struct {
	Clientset kubernetes.Interface
}

func (m *MockScanner) PodDetails() (*v1.PodList, error) {
	fakeClientSet := fake.NewClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-1",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				PodIP: "192.168.2.2",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "app",
						Ports: []v1.ContainerPort{
							{ContainerPort: 1000},
						},
					},
				},
			},
		},
	)
	return fakeClientSet.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
}

type Target struct {
	PodName string
	IP      string
	Port    int
}

type ScanResult struct {
	PodName   string `json:"PodName"`
	IP        string `json:"IP"`
	Port      int    `json:"Port"`
	Timestamp string `json:"Timestamp"`
}

type K8sScanner interface {
	PodDetails() (*v1.PodList, error)
}

type ClusterClient struct {
	Clientset kubernetes.Interface
}

func (c *ClusterClient) PodDetails() (*v1.PodList, error) {
	return c.Clientset.CoreV1().Pods(Namespace).List(context.Background(), metav1.ListOptions{})
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan K8s Pods for open ports",
	Long:  `Discovers Pod IPs in the cluster and scans them using a concurrent worker pool.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := AuthPodDetails()
		if err != nil {
			fmt.Printf("Error connecting to cluster: %v\n", err)
			os.Exit(1)
		}
		startScanner(client)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func AuthPodDetails() (K8sScanner, error) {
	userHomeDir, _ := os.UserHomeDir()
	kubeconfigPath := filepath.Join(userHomeDir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ClusterClient{Clientset: clientset}, nil
}

func startScanner(k8s K8sScanner) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("📈 Metrics server active at :2112/metrics")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			fmt.Printf("Metrics server failed: %v\n", err)
		}
	}()
	jobs := make(chan Target, 100)
	results := make(chan ScanResult, 100)
	var wg sync.WaitGroup
	start := time.Now()

	for w := 1; w <= 100; w++ {
		wg.Add(1)
		k8s_active_workers.Inc()
		go port_checker(&wg, jobs, results)
	}

	go func() {
		commonPorts := []int{80, 443, 8080}
		pods, err := k8s.PodDetails()
		if err != nil {
			fmt.Printf("Error fetching pods: %v\n", err)
			close(jobs)
			return
		}

		for _, pod := range pods.Items {
			scansTotal.Inc()
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
		scanDuration.Observe(time.Since(start).Seconds())
	}()

	var allResults []ScanResult
	fmt.Printf("🔍 Scanning namespace: %s...\n", Namespace)
	for res := range results {
		err := checkValidJSON(res)
		if err != nil {
			panic(err)
		}
		allResults = append(allResults, res)
	}
	jsonData, err := json.MarshalIndent(allResults, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results: %s\n", err.Error())
		return
	}

	fmt.Println(string(jsonData))
	select {} // blocks forever
}

func checkValidJSON(res ScanResult) error {
	if res.PodName == "" {
		return fmt.Errorf("PodName is empty")
	}
	if res.IP == "" {
		return fmt.Errorf("IP is empty")
	}
	if res.Port <= 0 {
		return fmt.Errorf("Port must be > 0")
	}
	return nil
}

func port_checker(wg *sync.WaitGroup, jobs <-chan Target, results chan<- ScanResult) {
	defer wg.Done()
	defer k8s_active_workers.Dec()
	for target := range jobs {

		address := fmt.Sprintf("%s:%d", target.IP, target.Port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			openPortsFound.Inc()
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
