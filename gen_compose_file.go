package main

import (
	"fmt"
	"os"
	"strconv"
)

func writeNflowBaseServiceBlock(f *os.File, config ConfigFile, configFile string) {
	f.WriteString("  nflow-base:\n")
	f.WriteString("    image: manflow\n")
	f.WriteString("    environment:\n")
	f.WriteString("      COLLECTOR_IP: " + config.CollectorIp + "\n")
	f.WriteString("      COLLECTOR_PORT: " + strconv.Itoa(config.CollectorPort) + "\n")
	f.WriteString("      CONFIG_FILE: config.json\n")
	f.WriteString("    volumes:\n")
	f.WriteString("      - ./" + configFile + ":/app/config.json\n")
	f.WriteString("    profiles:\n")
	f.WriteString("      - nflow-base\n")
	f.WriteString("\n")
}

func writeNflowServiceBlock(f *os.File, hostConfig ConfigHost) {
	f.WriteString("  nflow-" + hostConfig.Name + ":\n")
	f.WriteString("    extends:\n")
	f.WriteString("      service: nflow-base\n")
	f.WriteString("    environment:\n")
	f.WriteString("      HOST_NAME: " + hostConfig.Name + "\n")
	f.WriteString("    networks:\n")
	f.WriteString("      nflow-network:\n")
	f.WriteString("        ipv4_address: " + hostConfig.Ip + "\n")
	f.WriteString("    profiles:\n")
	f.WriteString("      - nflow\n")
	f.WriteString("\n")
}

func writeNflowNetworkBlock(f *os.File, config ConfigFile) {
	f.WriteString("networks:\n")
	f.WriteString("  nflow-network:\n")
	f.WriteString("    ipam:\n")
	f.WriteString("      driver: default\n")
	f.WriteString("      config:\n")
	f.WriteString("        - subnet: 10.0.0.0/16\n")
	f.WriteString("\n")
}

func GenComposeFile(filename string, config ConfigFile) error {
	composeFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to create compose file %s: %v", filename, err)
	}

	defer composeFile.Close()

	composeFile.WriteString("version: '5.3'\n\nservices:\n")

	writeNflowBaseServiceBlock(composeFile, config, filename)

	for _, hostConfig := range config.Hosts {
		writeNflowServiceBlock(composeFile, hostConfig)
	}

	writeNflowNetworkBlock(composeFile, config)

	err = composeFile.Sync()

	if err != nil {
		return fmt.Errorf("failed to sync compose file %s: %v", filename, err)
	}

	fmt.Println("Successfully generated compose file: " + filename)

	return nil
}
