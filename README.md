# 🔍 K8s-Scan: Concurrent Port Recon Module

A high-performance CLI tool built in **Go** to scan Kubernetes Pods for open ports. This project demonstrates mastery of Go concurrency patterns and the `client-go` library, designed for security auditing and network discovery in cloud-native environments.

---

### ⚡ Key Features

* [cite_start]**God-Level Concurrency**: Uses a **Worker Pool** pattern (Goroutines & Channels) to scan hundreds of ports simultaneously without resource exhaustion. [cite: 103, 104, 106]
* [cite_start]**K8s Service Discovery**: Authenticates with clusters via `client-go` to dynamically fetch Pod IPs across any namespace. [cite: 756, 1258]
* [cite_start]**Professional CLI**: Built with the **Cobra** framework, providing a standardized UX similar to `kubectl`. [cite: 1849, 1857]
* [cite_start]**Safe Execution**: Implements `sync.WaitGroup` for proper lifecycle management and `net.DialTimeout` for network resiliency. [cite: 170, 449]

---

### 🏗️ Architecture

The tool follows a professional **Producer-Consumer** architecture:
1.  [cite_start]**Producer**: Fetches Pod IPs from the Kubernetes API. [cite: 107]
2.  [cite_start]**Job Queue**: A buffered channel distributes tasks to workers. [cite: 107]
3.  [cite_start]**Worker Pool**: A fixed number of workers (50+) consume jobs and execute port checks. [cite: 109, 113]
4.  [cite_start]**Results Aggregator**: Results are collected via a separate channel and printed to the UI. [cite: 108]

---

### 🚀 Getting Started

#### Prerequisites
* Go 1.20+
* [cite_start]A running Kubernetes cluster (Minikube/Kind) [cite: 1261]
* [cite_start]Configured `~/.kube/config` [cite: 804, 824]

#### Installation
```bash
go mod tidy
go build -o k8s-scan main.go