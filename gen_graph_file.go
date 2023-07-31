package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GenGraphFile(filename string, flowConfigs []ConfigFlow) error {
	csvFile, err := os.Create(filename)

	if err != nil {
		return fmt.Errorf("failed to create graph file %s: %v", filename, err)
	}

	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)

	header := []string{"from", "to"}
	err = csvWriter.Write(header)

	if err != nil {
		panic(err)
	}

	rowMap := map[string]bool{}

	for i := 0; i < len(flowConfigs); i++ {
		flowConfig := flowConfigs[i]

		for j := 0; j < len(flowConfig.Hops); j++ {
			if j == 0 {
				row := []string{flowConfig.SrcAddr + ":" + strconv.Itoa(int(flowConfig.SrcPort)), flowConfig.Hops[j]}
				rowKey := strings.Join(row, "-")
				if _, ok := rowMap[rowKey]; !ok {
					csvWriter.Write(row)
					rowMap[rowKey] = true
				}
			} else {
				row := []string{flowConfig.Hops[j-1], flowConfig.Hops[j]}
				rowKey := strings.Join(row, "-")
				if _, ok := rowMap[rowKey]; !ok {
					csvWriter.Write([]string{flowConfig.Hops[j-1], flowConfig.Hops[j]})
					rowMap[rowKey] = true
				}
			}

			if j == len(flowConfig.Hops)-1 {
				row := []string{flowConfig.Hops[j], flowConfig.DstAddr + ":" + strconv.Itoa(int(flowConfig.DstPort))}
				rowKey := strings.Join(row, "-")
				if _, ok := rowMap[rowKey]; !ok {
					csvWriter.Write([]string{flowConfig.Hops[j], flowConfig.DstAddr + ":" + strconv.Itoa(int(flowConfig.DstPort))})
					rowMap[rowKey] = true
				}
			}
		}
	}

	csvWriter.Flush()

	fmt.Println("Generated graph file: " + filename)

	return nil
}
