// Run using:
//
//	$ ./a.out
//
// go run nflow-generator.go nflow_logging.go nflow_payload.go  -t 172.16.86.138 -p 9995
// Or:
// go build
// ./nflow-generator -t <ip> -p <port>
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Proto int

const (
	FTP Proto = iota + 1
	SSH
	DNS
	HTTP
	HTTPS
	NTP
	SNMP
	IMAPS
	MYSQL
	HTTPS_ALT
	P2P
	BITTORRENT
)

const MAX_FLOWS_PER_RECORD = 30

const TICK_INTERVAL_MS = 1000

func main() {
	// Parse arguments from command line and environment variables
	configArgs, err := ParseConfigArgs()

	if err != nil {
		panic(err)
	}

	// Read flow configuration file
	var config ConfigFile

	err = ReadFlowConfigFile(&config, configArgs.ConfigFile)

	if err != nil {
		panic(err)
	}

	// For generating docker compose configuration file
	if opts.GenComposeFile != "" {
		err := GenComposeFile(opts.GenComposeFile, config)

		if err != nil {
			panic(err)
		}

		return
	}

	// For generating prometheus service discovery file
	if opts.GenTargetsFile != "" {
		err := GenTargetsFile(opts.GenTargetsFile, config)

		if err != nil {
			panic(err)
		}

		return
	}

	if configArgs.HostName == "" {
		panic(fmt.Errorf("host name not provided"))
	}

	if (config.CollectorIp == "" || config.CollectorPort == 0) && !opts.Simulate {
		panic(fmt.Errorf("collector ip/port not provided"))
	}

	hostName := configArgs.HostName
	fmt.Println("Host: " + hostName)

	// Initialize random number generator using seed value
	randGen := InitRandGen(config)

	// Parse the field values for each flow
	multiFlowConfigs := ParseUserFlows(&config)

	// Expand flows with multiple values into multiple flows
	expandedFlowConfigs := ExpandMultiFlows(multiFlowConfigs)

	flowConfigs := expandedFlowConfigs

	// Populate all missing fields using seeded randgen before sending
	//  so multiple generators will have the same values
	SeedFlows(flowConfigs, randGen, configArgs, config)

	// Filter flows for this host
	enabledFlows := FilterEnabledFlows(flowConfigs)

	// Generate graph file for topology visualization (WIP)
	if opts.GenGraphFile != "" {
		err := GenGraphFile(opts.GenGraphFile, flowConfigs)

		if err != nil {
			panic(err)
		}

		os.Exit(0)
	}

	// Initialize flow state for each flow
	configFlowStates := InitFlowState(enabledFlows)

	// Print all configured flows for this host
	if !opts.DisableLogging {
		for i := 0; i < len(enabledFlows); i++ {
			flowConfig := flowConfigs[enabledFlows[i].ConfigIndex]

			fmt.Printf(
				"%15s = %15s %5d -> %15s %5d [%3d] = %d (tick %d) (count %d)\n",
				hostName,
				flowConfig.SrcAddr,
				flowConfig.SrcPort,
				flowConfig.DstAddr,
				flowConfig.DstPort,
				flowConfig.Proto,
				flowConfig.Bytes,
				flowConfig.Tick,
				flowConfig.Count,
			)
		}
	}

	// Print flow configuration information
	numEnabledFlows := len(enabledFlows)

	fmt.Println("Number of user flow configs: " + strconv.Itoa(len(config.Flows)))
	fmt.Println("Number of flows configured: " + strconv.Itoa(len(flowConfigs)))
	fmt.Println("Number of active flows: " + strconv.Itoa(numEnabledFlows))
	fmt.Println("Expected average flow rate: " + strconv.Itoa(numEnabledFlows/config.FlowTimeout))

	if numEnabledFlows == 0 {
		fmt.Println("No flows configured for this host")
		return
	}

	// Initialize UDP connection to netflow collector
	var conn *net.UDPConn

	if !opts.Simulate {
		conn, err = InitUdpConn(config)

		if err != nil {
			panic(err)
		}
	}

	// Initialize prometheus metrics server
	go HandleMetricsServer()

	// Sleep until the start of the next 10 second interval
	//  so that multiple generators start at roughly the same time
	now := time.Now()
	diff := time.Duration(10-now.Second()%10) * time.Second

	fmt.Printf("Sleeping for %v\n", diff)
	time.Sleep(diff)

	// Flows are sent every TICK_INTERVAL_MS
	tick := 0
	skipped := true
	maxTick := config.FlowTimeout * 1000 / TICK_INTERVAL_MS

	fmt.Println("Sending flows...")
	for {

		// Initialize bytes value for this tick
		// Note we initialize bytes for all flows, even if they are not enabled
		//  so that the same values are used for all generators
		for i := 0; i < len(flowConfigs); i++ {
			flowConfigs[i].Bytes = GenBytesValue(randGen)
		}

		// Send flows for this tick
		for i := 0; i < len(enabledFlows); {
			records := []NetflowPayload{}

			// Calculate sytem uptime for this tick
			// This value is used in the netflow packet header
			uptime := CreateCalcUptime()

			// We send MAX_FLOWS_PER_RECORD netflow records per netflow packet
			//  so we use a nested loop to send all flows for this tick
			j := 0
			for ; j < MAX_FLOWS_PER_RECORD && i < len(enabledFlows); i, j = i+1, j+1 {
				flowConfig := flowConfigs[enabledFlows[i].ConfigIndex]

				// Check if the flow count has been reached
				if flowConfig.Count != 0 && configFlowStates[i].Count+1 > flowConfig.Count {
					continue
				}

				skipped = false

				if flowConfig.Tick != tick {
					continue
				}

				// If the flow has multiple hops, check if we should provide a value
				//  for the next hop field
				numHops := len(flowConfig.Hops)

				nextHopHostName := ""
				if flowConfig.HostIndex < numHops-1 {
					nextHopHostName = flowConfig.Hops[flowConfig.HostIndex+1]
				}

				// Create the netflow record
				payload := CreateCustomFlow(
					flowConfig.SrcAddr,
					flowConfig.SrcPort,
					flowConfig.DstAddr,
					flowConfig.DstPort,
					flowConfig.Proto,
					FindHostIp(config.Hosts, nextHopHostName),
					flowConfig.Bytes,
					// TODO improve the logic for first_switched and last_switched
					int(TICK_INTERVAL_MS/numHops)*(numHops-flowConfig.HostIndex),
					int(TICK_INTERVAL_MS/numHops)*(numHops-flowConfig.HostIndex-1),
				)

				// Update the flow state
				configFlowStates[i].Count++
				configFlowStates[i].Bytes += flowConfig.Bytes

				// Print the flow record
				if !opts.DisableLogging {
					fmt.Printf(
						"%15s = %15s %5d -> %15s %5d [%3d] = %s -> %s = %d\n",
						hostName,
						ConvertIntToIp(payload.SrcIP).String(),
						payload.SrcPort,
						ConvertIntToIp(payload.DstIP).String(),
						payload.DstPort,
						payload.IpProtocol,
						time.Unix(int64(uptime.UnixSec+payload.SysUptimeStart/1000), int64(uptime.UnixMsec)).Format("2006-01-02T15:04:05.000Z"),
						time.Unix(int64(uptime.UnixSec+payload.SysUptimeEnd/1000), int64(uptime.UnixMsec)).Format("2006-01-02T15:04:05.000Z"),
						payload.NumOctets,
					)
				}

				records = append(records, payload)
			}

			if len(records) == 0 {
				continue
			}

			// Create the netflow packet
			data := new(Netflow)

			data.Header = CreateNFlowHeader(len(records))

			data.Records = records

			buffer := BuildNFlowPayload(*data)

			// Write the netflow packet to the UDP connection
			if !opts.Simulate {
				bytesWritten, err := conn.Write(buffer.Bytes())
				if err != nil {
					log.Fatal("Failed to write: ", err)
				}

				sentRecordsTotalBytesCounter.Add(float64(bytesWritten))
			}

			// Update prometheus metrics
			sentNetflowTotalCounter.Inc()
			sentRecordsTotalCounter.Add(float64(len(records)))
		}

		tick++
		if tick == maxTick {
			// If we went through a whole tick cycle without sending any flows
			//  then we are done sending flows
			if skipped {
				fmt.Println("No more flows to send")
				break
			}

			tick = 0
			skipped = true
		}

		// Sleep until the next tick
		sleepInt := time.Duration(TICK_INTERVAL_MS)
		time.Sleep(sleepInt * time.Millisecond)
	}

	if !opts.DisableLogging {
		fmt.Println("Done sending flows, here are the stats:")

		for i := 0; i < len(configFlowStates); i++ {
			flowConfig := flowConfigs[enabledFlows[i].ConfigIndex]
			flowConfigState := configFlowStates[i]

			fmt.Printf(
				"%15s = %15s %5d -> %15s %5d [%3d] = %d total = %d bytes\n",
				hostName,
				flowConfig.SrcAddr,
				flowConfig.SrcPort,
				flowConfig.DstAddr,
				flowConfig.DstPort,
				flowConfig.Proto,
				flowConfigState.Count,
				flowConfigState.Bytes,
			)
		}
	}

	if opts.StatsOutFile != "" {
		err := GenStatsFile(opts.StatsOutFile, configFlowStates, enabledFlows, flowConfigs)

		if err != nil {
			panic(err)
		}
	}
}
