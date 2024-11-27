# fl-orchestrator

## Prerequisites

1. Install Go version 1.20 or higher (https://go.dev/doc/install).

2. Deploy a Kubernetes (or K3s) cluster.

3. Edit `k8sConfigFilePath` in `"cmd/http/main.go"` line 41 to match path to cluster config.

## Deployment

### Actual infrastructure

1. Define FL node type (GA, LA, client) for each Kubernetes node through labels, like in `"scripts/set_labels.sh"`. Additionally, you can define communication cost between nodes and client node data distribution for dataset CIFAR-10.

2. Start fl-orchestrator:

```bash
cd cmd/http
go run main.go
```

### Simulated infrastructure

Simulated infrastructure means that it is defined in CSV files and not the actual infrastructure of Kubernetes cluster. With this option you can simulate multiple FL nodes on a single Kubernetes node to test your configurations and solutions.

For example, to reproduce experiment 1.a (`"experiments/icmlcn/1a"`):

1. Copy cluster configuration (`cluster.csv` and `changes.csv`) to `"configs/cluster"`: 

```bash
yes | cp -rf experiments/icmlcn/1a/cluster.csv configs/cluster/cluster.csv
yes | cp -rf experiments/icmlcn/1a/changes.csv configs/cluster/changes.csv
```

The given example above will configure experiment 1.a which trains a CNN model for an image classification task on the CIFAR-10 dataset. The experiment starts with the environment defined in `cluster.csv` (1 GA, 2 LAs, 8 clients with small IID datasets) and after round 10 it appends the content of `changes.csv` to `cluster.csv` to trigger reconfiguration.

2. Start fl-orchestrator as a simulation:

```bash
cd cmd/http
go run main.go sim
```

## Usage

To start an HFL task, send a POST request (with curl or Postman) to `http://<NODE_IP>:8080/fl/start`. Example of a request:

```json
{
    "epochs": 2,
    "localRounds": 2,
    "configurationModel": "minCommCost",
    "modelSize": 3.3, // 3.35 
    "costConfiguration": {
        "costType": "totalBudget",
        "communicationBudget": 100000
    },
    "rvaEnabled": true
}
```

This deploys HFL where local epochs is set to 2, local rounds also set to 2, and using a configuration strategy "minCommCost" that clusters clients to minimize cost between clients and local aggregators. It defines the total communication budget to be 100 000 units and enables RVA to be used by the orchestrator.
