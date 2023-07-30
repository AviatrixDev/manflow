package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PromTargetLabels struct {
	Ip   string `json:"ip"`
	Name string `json:"name"`
}

type PromTarget struct {
	Targets []string         `json:"targets"`
	Labels  PromTargetLabels `json:"labels"`
}

func GenTargetsFile(filename string, config ConfigFile) error {
	targetsFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to create targets file %s: %v", filename, err)
	}

	defer targetsFile.Close()

	var targets []PromTarget

	for _, hostConfig := range config.Hosts {
		targets = append(targets, PromTarget{
			Targets: []string{hostConfig.Ip + ":" + PROM_PORT},
			Labels: PromTargetLabels{
				Ip:   hostConfig.Ip,
				Name: hostConfig.Name,
			},
		})
	}

	targetsJson, err := json.MarshalIndent(targets, "", "\t")

	if err != nil {
		return fmt.Errorf("failed to marshal targets: %v", err)
	}

	targetsFile.WriteString(string(targetsJson))

	err = targetsFile.Sync()

	if err != nil {
		return fmt.Errorf("failed to sync targets file %s: %v", filename, err)
	}

	fmt.Println("Successfully generated targets file: " + opts.GenTargetsFile)

	return nil
}
