package florch

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
)

func BuildGlobalAggregatorConfigFiles(flAggregator *model.FlAggregator) (map[string]string, error) {
	configDirectoryPath := "../../configs/fl/"

	taskBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "task/task.py"))
	if err != nil {
		fmt.Print(err)
	}
	taskString := string(taskBytesArray)

	globalAggregatorConfig := GlobalAggregatorConfig_Yaml

	filesData := map[string]string{
		"task.py":                   taskString,
		"global_server_config.yaml": globalAggregatorConfig,
	}

	return filesData, nil
}

func BuildLocalAggregatorConfigFiles(flAggregator *model.FlAggregator) (map[string]string, error) {
	localAggregatorConfig := fmt.Sprintf(LocalAggregatorConfig_Yaml, flAggregator.ParentAddress, strconv.Itoa(int(flAggregator.LocalRounds)))

	filesData := map[string]string{
		"local_server_config.yaml": localAggregatorConfig,
	}

	return filesData, nil
}

func BuildClientConfigFiles(client *model.FlClient) (map[string]string, error) {
	configDirectoryPath := "../../configs/fl/"

	taskBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "task/task.py"))
	if err != nil {
		fmt.Print(err)
	}
	taskString := string(taskBytesArray)

	clientConfigString := fmt.Sprintf(ClientConfig_Yaml, client.ParentAddress, strconv.Itoa(int(client.PartitionId)),
		"10", strconv.Itoa(int(client.Epochs)))

	filesData := map[string]string{
		"task.py":            taskString,
		"client_config.yaml": clientConfigString,
	}

	return filesData, nil
}

const GlobalAggregatorConfig_Yaml = `
server:
  address: "0.0.0.0:8080"
  global_rounds: 100

strategy:
  fraction_fit: 1.0
  fraction_evaluate: 1.0
  min_fit_clients: 2
  min_evaluate_clients: 2
  min_available_clients: 2
`

const LocalAggregatorConfig_Yaml = `
server:
  global_address: "%[1]s"
  local_address: "0.0.0.0:8080"
  local_rounds: %[2]s
  global_rounds: 100

strategy:
  fraction_fit: 1.0
  fraction_evaluate: 1.0
  min_fit_clients: 2
  min_evaluate_clients: 2
  min_available_clients: 2
`

const ClientConfig_Yaml = `
server:
  address: "%[1]s"

node_config:
  partition-id: %[2]s 
  num-partitions: %[3]s 

run_config:
  batch-size: 32
  local-epochs: %[4]s 
  learning-rate: 0.1 
`
