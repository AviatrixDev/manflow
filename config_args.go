package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Help           bool   `short:"h" long:"help" description:"show nflow-generator help"`
	HostName       string `short:"i" long:"host-name" description:"provide host name to use with config file"`
	ConfigFile     string `short:"e" long:"config-file" description:"provide config file to describe complex flow generation behavior"`
	GenGraphFile   string `short:"g" long:"gen-graph-file" description:"generate graph file"`
	DisableLogging bool   `short:"l" long:"disable-logging" description:"disable logging"`
	Simulate       bool   `short:"m" long:"simulate" description:"simulate only, do not send to collector"`
	StatsOutFile   string `short:"o" long:"stats-out-file" description:"write stats to file"`
	GenComposeFile string `short:"q" long:"gen-compose-file" description:"generate compose file"`
	GenTargetsFile string `short:"r" long:"gen-targets-file" description:"generate prometheus targets file"`
}

type ConfigArgs struct {
	ConfigFile string
	HostName   string
}

func ParseConfigArgs() (ConfigArgs, error) {
	_, err := flags.Parse(&opts)

	if err != nil {
		return ConfigArgs{}, fmt.Errorf("failed to parse config args: %v", err)
	}

	inputConfigFile := "flowConfig.json"
	if opts.ConfigFile != "" {
		inputConfigFile = opts.ConfigFile
	} else if os.Getenv("CONFIG_FILE") != "" {
		inputConfigFile = os.Getenv("CONFIG_FILE")
	}

	inputHostName := ""
	if opts.HostName != "" {
		inputHostName = opts.HostName
	} else if os.Getenv("HOST_NAME") != "" {
		inputHostName = os.Getenv("HOST_NAME")
	}

	return ConfigArgs{
		ConfigFile: inputConfigFile,
		HostName:   inputHostName,
	}, nil
}

func showUsage() {
	var usage string
	usage = `
Usage:
  main [OPTIONS] [collector IP address] [collector port number]

  Send mock Netflow version 5 data to designated collector IP & port.
  Time stamps in all datagrams are set to UTC.

Application Options:
  -t, --target= target ip address of the netflow collector
  -p, --port=   port number of the target netflow collector
  -s, --spike run a second thread generating a spike for the specified protocol
    protocol options are as follows:
        ftp - generates tcp/21
        ssh  - generates tcp/22
        dns - generates udp/54
        http - generates tcp/80
        https - generates tcp/443
        ntp - generates udp/123
        snmp - generates ufp/161
        imaps - generates tcp/993
        mysql - generates tcp/3306
        https_alt - generates tcp/8080
        p2p - generates udp/6681
        bittorrent - generates udp/6682
  -f, --false-index generate a false snmp index values of 1 or 2. The default is 0. (Optional)
  -c, --flow-count set the number of flows to generate in each iteration. The default is 16. (Optional)

Example Usage:

    -first build from source (one time)
    go build   

    -generate default flows to device 172.16.86.138, port 9995
    ./nflow-generator -t 172.16.86.138 -p 9995 

    -generate default flows along with a spike in the specified protocol:
    ./nflow-generator -t 172.16.86.138 -p 9995 -s ssh

    -generate default flows with "false index" settings for snmp interfaces 
    ./nflow-generator -t 172.16.86.138 -p 9995 -f

    -generate default flows with up to 256 flows
    ./nflow-generator -c 128 -t 172.16.86.138 -p 9995

Help Options:
  -h, --help    Show this help message
  `
	fmt.Print(usage)
}
