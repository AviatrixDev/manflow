package main

import (
	"fmt"
	"net"
	"strconv"
)

func InitUdpConn(config ConfigFile) (*net.UDPConn, error) {
	collector := config.CollectorIp + ":" + strconv.Itoa(config.CollectorPort)

	udpAddr, err := net.ResolveUDPAddr("udp", collector)

	if err != nil {
		return nil, fmt.Errorf("failed to resolve udp addr %s: %v", collector, err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		return nil, fmt.Errorf("failed to dial udp addr %s: %v", collector, err)
	}

	return conn, nil
}
