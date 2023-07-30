package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
)

func FindIndex(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}

	return -1
}

func ConvertRangeToSlice(start, end int) []int {
	var result []int
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result
}

func ConvertPortRangeToSlice(start, end int) []uint16 {
	var result []uint16
	for i := start; i <= end; i++ {
		result = append(result, uint16(i))
	}
	return result
}

func GetCidrHosts(cidr string) ([]string, error) {
	prefix, err := netip.ParsePrefix(cidr)

	if err != nil {
		return nil, fmt.Errorf("failed to parse cidr %s: %v", cidr, err)
	}

	var ips []string
	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {
		ips = append(ips, addr.String())
	}

	if len(ips) < 2 {
		return ips, nil
	}

	return ips[1 : len(ips)-1], nil
}

func ConvertIntToIp(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func FindHostIp(hosts []ConfigHost, name string) string {
	if name == "" {
		return ""
	}

	for _, host := range hosts {
		if host.Name == name {
			return host.Ip
		}
	}
	panic("host not found: " + name)
}

func randomNum(min, max int) int {
	return rand.Intn(max-min) + min
}
