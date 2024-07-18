package florch

import (
	"fmt"
	"math/rand"
	"os"

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

	loggingConfig := buildLoggingConfig(fmt.Sprintf("ga-%s", flAggregator.Id))
	loggingConfigString, err := interfaceToYamlString(loggingConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	loggingConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/logging_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	loggingConfigString = fmt.Sprintf("%s\n%s", loggingConfigString, string(loggingConfigBytesArray))

	modelConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/model_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	modelConfigString := string(modelConfigBytesArray)

	globalAggregatorConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "global_aggregator/aggregator_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	globalAggregatorConfigString := string(globalAggregatorConfigBytesArray)

	globalAggregatorEntryConfig := buildGlobalAggregatorEntryConfigVarying(flAggregator)
	globalAggregatorEntryConfigString, err := interfaceToYamlString(globalAggregatorEntryConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	globalAggregatorEntryConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "global_aggregator/entry.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	globalAggregatorEntryConfigString = fmt.Sprintf("%s\n%s", globalAggregatorEntryConfigString, string(globalAggregatorEntryConfigBytesArray))

	filesData := map[string]string{
		"entry.yaml":             globalAggregatorEntryConfigString,
		"aggregator_config.yaml": globalAggregatorConfigString,
		"logging_config.yaml":    loggingConfigString,
		"datasets_config.yaml":   datasetsConfigString,
		"model_config.yaml":      modelConfigString,
	}

	return filesData, nil
}

func BuildLocalAggregatorConfigFiles(flAggregator *model.FlAggregator) (map[string]string, error) {
	configDirectoryPath := "../../configs/fl/"

	datasetsConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/datasets_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	datasetsConfigString := string(datasetsConfigBytesArray)

	loggingConfig := buildLoggingConfig(fmt.Sprintf("la-%s", flAggregator.Id))
	loggingConfigString, err := interfaceToYamlString(loggingConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	loggingConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/logging_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	loggingConfigString = fmt.Sprintf("%s\n%s", loggingConfigString, string(loggingConfigBytesArray))

	modelConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/model_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	modelConfigString := string(modelConfigBytesArray)

	localAggregatorConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "local_aggregator/aggregator_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	localAggregatorConfigString := string(localAggregatorConfigBytesArray)

	localAggregatorEntryConfig := buildLocalAggregatorEntryConfigVarying(flAggregator)
	localAggregatorEntryConfigString, err := interfaceToYamlString(localAggregatorEntryConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	localAggregatorEntryConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "local_aggregator/entry.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	localAggregatorEntryConfigString = fmt.Sprintf("%s\n%s", localAggregatorEntryConfigString, string(localAggregatorEntryConfigBytesArray))

	filesData := map[string]string{
		"entry.yaml":             localAggregatorEntryConfigString,
		"aggregator_config.yaml": localAggregatorConfigString,
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

	loggingConfig := buildLoggingConfig(fmt.Sprintf("cl-%s", client.Id))
	loggingConfigString, err := interfaceToYamlString(loggingConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	loggingConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/logging_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	loggingConfigString = fmt.Sprintf("%s\n%s", loggingConfigString, string(loggingConfigBytesArray))

	modelConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "shared/model_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	modelConfigString := string(modelConfigBytesArray)

	clientConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "client/client_config.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	clientConfigString := string(clientConfigBytesArray)

	clientEntryConfig := buildClientEntryConfigVarying(client)
	clientEntryConfigString, err := interfaceToYamlString(clientEntryConfig)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	clientEntryConfigBytesArray, err := os.ReadFile(fmt.Sprint(configDirectoryPath, "client/entry.yaml"))
	if err != nil {
		fmt.Print(err)
	}
	clientEntryConfigString = fmt.Sprintf("%s\n%s", clientEntryConfigString, string(clientEntryConfigBytesArray))

	filesData := map[string]string{
		"entry.yaml":           clientEntryConfigString,
		"client_config.yaml":   clientConfigString,
		"logging_config.yaml":  loggingConfigString,
		"datasets_config.yaml": datasetsConfigString,
		"model_config.yaml":    modelConfigString,
	}

	return filesData, nil
}

func buildGlobalAggregatorEntryConfigVarying(flAggregator *model.FlAggregator) model.GlobalAggregatorEntryConfig {
	return model.GlobalAggregatorEntryConfig{
		NumClients:    flAggregator.NumClients,
		Rounds:        flAggregator.Rounds,
		ServerAddress: flAggregator.InternalAddress,
	}
}

func buildLocalAggregatorEntryConfigVarying(flAggregator *model.FlAggregator) model.LocalAggregatorEntryConfig {
	return model.LocalAggregatorEntryConfig{
		NumClients:    flAggregator.NumClients,
		ServerAddress: flAggregator.InternalAddress,
		ParentAddress: flAggregator.ParentAddress,
		Rounds:        flAggregator.Rounds,
		LocalRounds:   flAggregator.LocalRounds,
	}
}

func buildClientEntryConfigVarying(client *model.FlClient) model.ClientEntryConfig {
	return model.ClientEntryConfig{
		//ClientId:      client.Id,
		ClientId:      int32(rand.Intn(9) + 1),
		Epochs:        client.Epochs,
		ServerAddress: client.ParentAddress,
	}
}

func buildLoggingConfig(runName string) model.LoggingConfig {
	return model.LoggingConfig{
		RunName: runName,
	}
}

func interfaceToYamlString(i interface{}) (string, error) {
	yamlData, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}
