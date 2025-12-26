# 🚀 Kubernetes Port Scanner (k8s-scan)

A high-performance CLI tool written in **Go** designed to audit and discover open ports across Kubernetes Pods. By leveraging Go's powerful concurrency primitives and the official `client-go` library, `k8s-scan` identifies network vulnerabilities and exposed services within your cluster at scale.

---

## ✨ Key Features

* **⚡ High-Performance Concurrency**: Utilizes a robust **Worker Pool** pattern with 50 concurrent goroutines to scan hundreds of targets simultaneously without overwhelming system resources.
* **☸️ Native K8s Integration**: Directly communicates with the Kubernetes API to dynamically discover Pod IPs in real-time.
* **🛠️ Professional CLI Interface**: Built on the **Cobra** framework, providing a familiar, `kubectl`-like experience with flags and subcommands.
* **🛡️ Resilient Scanning**: Implements `net.DialTimeout` to ensure the scanner doesn't hang on unresponsive network segments.

---

## 🏗️ How It Works

The tool implements a **Producer-Consumer** architecture:

1. **Discovery (Producer)**: The tool fetches all Pods from a specified namespace using your local `~/.kube/config`.
2. **Job Distribution**: Pod IPs and target ports (80, 443, 8080) are pushed into a buffered "jobs" channel.
3. **Worker Pool (Consumer)**: 50 workers pull tasks from the channel and attempt to establish TCP connections.
4. **Result Aggregation**: Successes are collected via a results channel and printed to the console in real-time.

---

## 🚦 Prerequisites

Before running the tool, ensure you have:

* **Go**: Version 1.25.5 or higher.
* **Kubeconfig**: A valid configuration file located at `~/.kube/config`.
* **Permissions**: RBAC permissions to `list` pods in the target namespace.

---

## 📥 Installation

Clone the repository and build the binary:

```bash
# Build the project
go build -o k8s-scan main.go

# (Optional) Move to your path
mv k8s-scan /usr/local/bin/

```

---

## 🚀 Usage

The scanner defaults to the `default` namespace but can be targeted at any specific namespace using flags.

### Scan Default Namespace

```bash
./k8s-scan scan

```

### Scan a Specific Namespace

```bash
./k8s-scan scan --namespace my-app-production
# OR
./k8s-scan scan -n kube-system

```

---

## 🛠️ Configuration & Dependencies

This project relies on professional-grade Go modules:

* `k8s.io/client-go`: For Kubernetes cluster communication.
* `github.com/spf13/cobra`: For the CLI structure and flag management.
* `k8s.io/apimachinery`: For Kubernetes API schema definitions.

---

## 📝 License

Copyright © 2025. All rights reserved.
