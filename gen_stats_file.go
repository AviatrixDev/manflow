package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigFlowState struct {
	Count int
	Bytes int
}

func InitFlowState(enabledFlows []EnabledConfigFlow) []ConfigFlowState {
	var configFlowStates []ConfigFlowState

	for i := 0; i < len(enabledFlows); i++ {
		configFlowState := new(ConfigFlowState)
		configFlowState.Count = 0
		configFlowState.Bytes = 0
		configFlowStates = append(configFlowStates, *configFlowState)
	}

	return configFlowStates
}

type OutStats struct {
	Total []ConfigFlowState `json:"total"`
}

func GenStatsFile(filename string, configFlowStates []ConfigFlowState) error {
	statsFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to create stats file %s: %v", filename, err)
	}

	defer statsFile.Close()

	outStats := OutStats{
		Total: configFlowStates,
	}

	result, err := json.Marshal(outStats)

	if err != nil {
		return fmt.Errorf("failed to marshal stats: %v", err)
	}

	_, err = statsFile.Write(result)

	if err != nil {
		return fmt.Errorf("failed to write stats to file %s: %v", filename, err)
	}

	err = statsFile.Sync()

	if err != nil {
		return fmt.Errorf("failed to sync stats file %s: %v", filename, err)
	}

	fmt.Println("Successfully generated stats file: " + filename)

	return nil
}
