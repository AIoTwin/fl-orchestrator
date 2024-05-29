# fl-orchestrator

## Prerequisites

1. Install Go version 1.20 or higher (https://go.dev/doc/install).

2. Deploy a Kubernetes (or K3s) cluster.

3. Edit "cmd/k8s/main.go" to match path to cluster config.

## Usage

Example (model size 10, communication budget 600):

```bash
cd cmd/k8s
go run main.go 10 600
```
