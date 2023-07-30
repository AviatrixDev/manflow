package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ConfigHost struct {
	Ip   string `json:"ip"`
	Name string `json:"name"`
}

type ConfigFlowUser struct {
	SrcAddr string   `json:"src_addr"`
	SrcPort string   `json:"src_port"`
	DstAddr string   `json:"dst_addr"`
	DstPort string   `json:"dst_port"`
	Proto   string   `json:"proto"`
	Hops    []string `json:"hops"`
	Count   int      `json:"count"`
}

type ConfigFile struct {
	Seed          int              `json:"seed"`
	FlowTimeout   int              `json:"flow_timeout"`
	CollectorIp   string           `json:"collector_ip"`
	CollectorPort int              `json:"collector_port"`
	Hosts         []ConfigHost     `json:"hosts"`
	Flows         []ConfigFlowUser `json:"flows"`
}

func ReadFlowConfigFile(config *ConfigFile, filename string) error {
	jsonFile, err := os.Open(filename)

	if err != nil {
		return fmt.Errorf("failed to open config file %s: %v", filename, err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, config)

	if err != nil {
		return fmt.Errorf("failed to parse config file %s: %v", filename, err)
	}

	if config.FlowTimeout == 0 {
		config.FlowTimeout = 60
	}

	return nil
}
