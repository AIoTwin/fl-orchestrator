package florch

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/model"
	"gopkg.in/yaml.v2"
)

func BuildGlobalAggregatorConfigFiles(flAggregator *model.FlAggregator) (map[string]string, error) {
	configDirectoryPath := "../../configs/fl/"

	datasetsConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/datasets_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	datasetsConfigString := string(datasetsConfigBytesArray)

	loggingConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/logging_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	loggingConfigString := string(loggingConfigBytesArray)

	modelConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/model_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	modelConfigString := string(modelConfigBytesArray)

	globalAggregatorEntryConfig := buildGlobalAggregatorEntryConfigVarying(flAggregator)
	globalAggregatorEntryConfigString, err := interfaceToYamlString(globalAggregatorEntryConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	globalAggregatorEntryConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "global_aggregator/entry_server.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	globalAggregatorEntryConfigString = fmt.Sprintf("%s\n%s", globalAggregatorEntryConfigString, string(globalAggregatorEntryConfigBytesArray))

	globalAggregatorConfig := buildGlobalAggregatorConfigVarying(flAggregator)
	globalAggregatorConfigString, err := interfaceToYamlString(globalAggregatorConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	globalAggregatorConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "global_aggregator/aggregator_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	globalAggregatorConfigString = fmt.Sprintf("%s\n%s", globalAggregatorConfigString, string(globalAggregatorConfigBytesArray))

	filesData := map[string]string{
		"entry_server.yaml":      globalAggregatorEntryConfigString,
		"aggregator_config.yaml": globalAggregatorConfigString,
		"logging_config.yaml":    loggingConfigString,
		"datasets_config.yaml":   datasetsConfigString,
		"model_config.yaml":      modelConfigString,
	}

	return filesData, nil
}

func BuildClientConfigFiles(client *model.FlClient) (map[string]string, error) {
	configDirectoryPath := "../../configs/fl/"

	datasetsConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/datasets_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	datasetsConfigString := string(datasetsConfigBytesArray)

	loggingConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/logging_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	loggingConfigString := string(loggingConfigBytesArray)

	modelConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/model_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	modelConfigString := string(modelConfigBytesArray)

	clientEntryConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "client/entry_client.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	clientEntryConfigString := string(clientEntryConfigBytesArray)

	clientConfig := buildClientConfigVarying(client)
	clientConfigString, err := interfaceToYamlString(clientConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	clientConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "client/client_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	clientConfigString = fmt.Sprintf("%s\n%s", clientConfigString, string(clientConfigBytesArray))

	filesData := map[string]string{
		"entry_client.yaml":    clientEntryConfigString,
		"client_config.yaml":   clientConfigString,
		"logging_config.yaml":  loggingConfigString,
		"datasets_config.yaml": datasetsConfigString,
		"model_config.yaml":    modelConfigString,
	}

	return filesData, nil
}

func buildGlobalAggregatorConfigVarying(flAggregator *model.FlAggregator) model.GlobalAggregatorConfig {
	return model.GlobalAggregatorConfig{
		ServerAddress: flAggregator.InternalAddress,
		Rounds:        flAggregator.Rounds,
	}
}

func buildGlobalAggregatorEntryConfigVarying(flAggregator *model.FlAggregator) model.GlobalAggregatorEntryConfig {
	return model.GlobalAggregatorEntryConfig{
		NumClients: flAggregator.NumClients,
	}
}

func buildClientConfigVarying(client *model.FlClient) model.ClientConfig {
	return model.ClientConfig{
		//ClientId:      client.Id,
		ClientId:      strconv.Itoa(rand.Intn(9) + 1),
		ServerAddress: client.ParentAddress,
		Epochs:        client.Epochs,
	}
}

func interfaceToYamlString(i interface{}) (string, error) {
	yamlData, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}
