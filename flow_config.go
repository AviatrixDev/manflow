package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type ConfigFlow struct {
	SrcAddr   string
	SrcPort   uint16
	DstAddr   string
	DstPort   uint16
	Proto     int
	Bytes     int
	Hops      []string
	Count     int
	Tick      int
	HostIndex int
}

type EnabledConfigFlow struct {
	ConfigIndex int
}

type ConfigFlowMultiple struct {
	SrcAddr []string
	SrcPort []uint16
	DstAddr []string
	DstPort []uint16
	Proto   []int
	Hops    []string
	Count   int
}

func ParseUserIpInput(input string) []string {
	if strings.Contains(input, "/") {
		ips, err := GetCidrHosts(input)

		if err != nil {
			panic(err)
		}

		return ips
	} else {
		return []string{input}
	}
}

func ParseUserProtoInput(input string) []int {
	if strings.Contains(input, ",") {
		ports := strings.Split(input, ",")
		var portSlice []int
		for _, port := range ports {
			portInt, err := strconv.Atoi(port)

			if err != nil {
				panic(fmt.Errorf("failed to parse port %s: %v", port, err))
			}

			portSlice = append(portSlice, portInt)
		}
		return portSlice
	} else if strings.Contains(input, "-") {
		ports := strings.Split(input, "-")
		startPort, err := strconv.Atoi(ports[0])

		if err != nil {
			panic(fmt.Errorf("failed to parse port %s: %v", ports[0], err))
		}

		endPort, err := strconv.Atoi(ports[1])

		if err != nil {
			panic(fmt.Errorf("failed to parse port %s: %v", ports[1], err))
		}

		return ConvertRangeToSlice(startPort, endPort)
	} else if input == "" {
		return []int{0}
	} else {
		port, err := strconv.Atoi(input)

		if err != nil {
			panic(fmt.Errorf("failed to parse port %s: %v", input, err))
		}

		return []int{port}
	}
}

func ParseUserPortInput(input string) []uint16 {
	var portSlice []uint16

	intSlice := ParseUserProtoInput(input)

	for _, port := range intSlice {
		portSlice = append(portSlice, uint16(port))
	}

	return portSlice
}

func ParseUserFlows(config *ConfigFile) []ConfigFlowMultiple {
	var multiFlowConfigs []ConfigFlowMultiple

	// Convert single value into multi value
	for i := 0; i < len(config.Flows); i++ {
		flow := config.Flows[i]

		multiFlowConfig := new(ConfigFlowMultiple)

		multiFlowConfig.SrcAddr = ParseUserIpInput(flow.SrcAddr)
		multiFlowConfig.DstAddr = ParseUserIpInput(flow.DstAddr)

		multiFlowConfig.SrcPort = ParseUserPortInput(flow.SrcPort)
		multiFlowConfig.DstPort = ParseUserPortInput(flow.DstPort)

		multiFlowConfig.Proto = ParseUserProtoInput(flow.Proto)

		multiFlowConfig.Hops = flow.Hops
		multiFlowConfig.Count = flow.Count

		multiFlowConfigs = append(multiFlowConfigs, *multiFlowConfig)
	}

	return multiFlowConfigs
}

func SeedFlows(flowConfigs []ConfigFlow, randGen *rand.Rand, configArgs ConfigArgs, config ConfigFile) {
	for i := 0; i < len(flowConfigs); i++ {
		flowConfig := flowConfigs[i]

		if flowConfig.SrcAddr == "" {
			flowConfigs[i].SrcAddr = ConvertIntToIp(randGen.Uint32()).String()
		}
		if flowConfig.DstAddr == "" {
			flowConfigs[i].DstAddr = ConvertIntToIp(randGen.Uint32()).String()
		}
		if flowConfig.SrcPort == 0 {
			flowConfigs[i].SrcPort = uint16(randGen.Intn(65535))
		}
		if flowConfig.DstPort == 0 {
			flowConfigs[i].DstPort = uint16(randGen.Intn(65535))
		}
		if flowConfig.Proto == 0 {
			flowConfigs[i].Proto = 6
		}
		// Bytes are initialized during the sending of the flow

		flowConfigs[i].Tick = randGen.Intn(config.FlowTimeout)

		hostIndex := FindIndex(configArgs.HostName, flowConfig.Hops)

		flowConfigs[i].HostIndex = hostIndex
	}
}

func ExpandMultiFlows(multiFlowConfigs []ConfigFlowMultiple) []ConfigFlow {
	var expandedFlowConfigs []ConfigFlow

	for i := 0; i < len(multiFlowConfigs); i++ {
		for _, srcAddr := range multiFlowConfigs[i].SrcAddr {
			for _, dstAddr := range multiFlowConfigs[i].DstAddr {
				for _, srcPort := range multiFlowConfigs[i].SrcPort {
					for _, dstPort := range multiFlowConfigs[i].DstPort {
						for _, proto := range multiFlowConfigs[i].Proto {
							flow := new(ConfigFlow)
							flow.SrcAddr = srcAddr
							flow.DstAddr = dstAddr
							flow.SrcPort = srcPort
							flow.DstPort = dstPort
							flow.Proto = proto
							flow.Hops = multiFlowConfigs[i].Hops
							flow.Count = multiFlowConfigs[i].Count
							expandedFlowConfigs = append(expandedFlowConfigs, *flow)
						}
					}
				}
			}
		}
	}

	return expandedFlowConfigs
}

func FilterEnabledFlows(flowConfigs []ConfigFlow) []EnabledConfigFlow {
	var enabledFlows []EnabledConfigFlow

	for i := 0; i < len(flowConfigs); i++ {
		flowConfig := flowConfigs[i]

		if flowConfig.HostIndex != -1 {
			enabledFlow := new(EnabledConfigFlow)
			enabledFlow.ConfigIndex = i

			enabledFlows = append(enabledFlows, *enabledFlow)
		}
	}

	return enabledFlows
}
