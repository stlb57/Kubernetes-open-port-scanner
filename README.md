# 🚀 Kubernetes Port Scanner (k8s-scan)

A high-performance, **Instrumented CLI Sentinel** written in **Go**. Designed to audit and discover open ports across Kubernetes Pods at scale, `k8s-scan` provides real-time visibility into your cluster's network security posture.

---

## ✨ Key Features

* **⚡ High-Performance Concurrency**: Utilizes a robust **Worker Pool** pattern with 100 concurrent goroutines to scan targets with maximum throughput.
* **📈 Industrial Observability**: Integrated **Prometheus Exporter** exposing real-time system "vitals" via a `/metrics` endpoint.
* **☸️ Native K8s Integration**: Directly communicates with the Kubernetes API to dynamically discover Pod IPs in real-time.
* **📊 Structured JSON Reporting**: Generates machine-readable, pretty-printed JSON audit reports for forensic analysis.
* **🛠️ Professional CLI Interface**: Built on the **Cobra** framework for a familiar, `kubectl`-like experience.

---

## 🏗️ How It Works

The tool implements a **Producer-Consumer** architecture enhanced with an **Observability Layer**:

1. **Discovery (Producer)**: Fetches Pods from the target namespace and increments the `k8s_port_scans_total` counter for every pod identified.
2. **Job Distribution**: Targets are fed into a buffered channel.
3. **Worker Pool (Consumer)**: 100 workers execute the scan. The `k8s_active_workers` gauge tracks real-time pool load, while the `k8s_open_ports_total` counter records every discovery.
4. **Metrics Server**: A background HTTP server exposes all internal metrics at `:2112/metrics` for Prometheus scraping.
5. **Result Aggregation**: Successes are collected and exported as a structured JSON report.

---

## 📊 Observability & Metrics

To achieve the **"Visual Truth,"** this tool exposes a Prometheus-compatible endpoint.

### Exposed Metrics:

| Metric Name | Type | Description |
| --- | --- | --- |
| `k8s_port_scans_total` | Counter | Total volume of pods identified for scanning. |
| `k8s_open_ports_total` | Counter | Total number of security "hits" (open ports) found. |
| `k8s_active_workers` | Gauge | Number of concurrent workers currently active in the pool. |
| `k8s_scan_duration_seconds` | Histogram | Latency distribution of the total scan operation. |

**Accessing Metrics:**

```bash
curl http://localhost:2112/metrics

```

---

## 📋 Data Model

The scanner uses a structured `ScanResult` object with UTC timestamps:

```go
type ScanResult struct {
	PodName   string `json:"PodName"`
	IP        string `json:"IP"`
	Port      int    `json:"Port"`
	Timestamp string `json:"Timestamp"`
}

```

---

## 🚦 Prerequisites

* **Go**: Version 1.25.5 or higher.
* **Kubeconfig**: Valid config at `~/.kube/config`.
* **Prometheus**: (Optional) For scraping the metrics endpoint.

---

## 🚀 Usage

### Run Scan & Expose Metrics

```bash
go run main.go scan

```

### Scraping with Prometheus

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'k8s-scanner'
    static_configs:
      - targets: ['localhost:2112']

```

---

## 📝 License

Copyright © 2026. All rights reserved.
