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

type OutStatsTotal struct {
	Count   int    `json:"count"`
	Bytes   int    `json:"bytes"`
	SrcAddr string `json:"src_addr"`
	SrcPort uint16 `json:"src_port"`
	DstAddr string `json:"dst_addr"`
	DstPort uint16 `json:"dst_port"`
	Proto   int    `json:"proto"`
}

type OutStats struct {
	Total []OutStatsTotal `json:"total"`
}

func GenStatsFile(filename string, configFlowStates []ConfigFlowState, enabledFlows []EnabledConfigFlow, flowConfigs []ConfigFlow) error {
	statsFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to create stats file %s: %v", filename, err)
	}

	defer statsFile.Close()

	var outStatsTotal []OutStatsTotal

	for i := 0; i < len(configFlowStates); i++ {
		flowState := configFlowStates[i]
		flowConfig := flowConfigs[enabledFlows[i].ConfigIndex]

		outStatsTotal = append(outStatsTotal, OutStatsTotal{
			Count:   flowState.Count,
			Bytes:   flowState.Bytes,
			SrcAddr: flowConfig.SrcAddr,
			SrcPort: flowConfig.SrcPort,
			DstAddr: flowConfig.DstAddr,
			DstPort: flowConfig.DstPort,
			Proto:   flowConfig.Proto,
		})
	}

	outStats := OutStats{
		Total: outStatsTotal,
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
