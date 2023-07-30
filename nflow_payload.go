package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"time"
)

// Start time for this instance, used to compute sysUptime
var StartTime = time.Now().UnixNano()

// current sysUptime in msec - recalculated in CreateNFlowHeader()
var sysUptime uint32 = 0

// Counter of flow packets that have been sent
var flowSequence uint32 = 0

const (
	FTP_PORT        = 21
	SSH_PORT        = 22
	DNS_PORT        = 53
	HTTP_PORT       = 80
	HTTPS_PORT      = 443
	NTP_PORT        = 123
	SNMP_PORT       = 161
	IMAPS_PORT      = 993
	MYSQL_PORT      = 3306
	HTTPS_ALT_PORT  = 8080
	P2P_PORT        = 6681
	BITTORRENT_PORT = 6682
	UINT16_MAX      = 65535
	PAYLOAD_AVG_MD  = 1024
	PAYLOAD_AVG_SM  = 256
)

// struct data from fach
type NetflowHeader struct {
	Version        uint16
	FlowCount      uint16
	SysUptime      uint32
	UnixSec        uint32
	UnixMsec       uint32
	FlowSequence   uint32
	EngineType     uint8
	EngineId       uint8
	SampleInterval uint16
}

type NetflowPayload struct {
	SrcIP          uint32
	DstIP          uint32
	NextHopIP      uint32
	SnmpInIndex    uint16
	SnmpOutIndex   uint16
	NumPackets     uint32
	NumOctets      uint32
	SysUptimeStart uint32
	SysUptimeEnd   uint32
	SrcPort        uint16
	DstPort        uint16
	Padding1       uint8
	TcpFlags       uint8
	IpProtocol     uint8
	IpTos          uint8
	SrcAsNumber    uint16
	DstAsNumber    uint16
	SrcPrefixMask  uint8
	DstPrefixMask  uint8
	Padding2       uint16
}

// Complete netflow records
type Netflow struct {
	Header  NetflowHeader
	Records []NetflowPayload
}

// Marshall NetflowData into a buffer
func BuildNFlowPayload(data Netflow) bytes.Buffer {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, &data.Header)
	if err != nil {
		log.Println("Writing netflow header failed:", err)
	}
	for _, record := range data.Records {
		err := binary.Write(buffer, binary.BigEndian, &record)
		if err != nil {
			log.Println("Writing netflow record failed:", err)
		}
	}
	return *buffer
}

type Uptime struct {
	UnixSec  uint32
	UnixMsec uint32
}

func CreateCalcUptime() Uptime {
	t := time.Now().UnixNano()
	sec := t / int64(time.Second)
	nsec := t - sec*int64(time.Second)
	sysUptime = uint32((t-StartTime)/int64(time.Millisecond)) + 1000

	uptime := new(Uptime)
	uptime.UnixSec = uint32(sec)
	uptime.UnixMsec = uint32(nsec)
	return *uptime
}

// Generate and initialize netflow header
func CreateNFlowHeader(recordCount int) NetflowHeader {

	t := time.Now().UnixNano()
	sec := t / int64(time.Second)
	nsec := t - sec*int64(time.Second)
	sysUptime = uint32((t-StartTime)/int64(time.Millisecond)) + 1000
	flowSequence++

	// log.Infof("Time: %d; Seconds: %d; Nanoseconds: %d\n", t, sec, nsec)
	// log.Infof("StartTime: %d; sysUptime: %d", StartTime, sysUptime)
	// log.Infof("FlowSequence %d", flowSequence)

	h := new(NetflowHeader)
	h.Version = 5
	h.FlowCount = uint16(recordCount)
	h.SysUptime = sysUptime
	h.UnixSec = uint32(sec)
	h.UnixMsec = uint32(nsec)
	h.FlowSequence = flowSequence
	h.EngineType = 1
	h.EngineId = 0
	h.SampleInterval = 0
	return *h
}

func CreateCustomFlow(
	srcIp string,
	srcPort uint16,
	dstIp string,
	dstPort uint16,
	protocol int,
	nextHopIp string,
	bytes int,
	startOffset int,
	endOffset int,
) NetflowPayload {
	payload := new(NetflowPayload)

	FillCommonFields(payload, PAYLOAD_AVG_SM, protocol, rand.Intn(32))

	offsetCenter := int((startOffset-endOffset)/2) + endOffset

	uptime := int(sysUptime)
	payload.SysUptimeEnd = uint32(uptime - randomNum(endOffset, offsetCenter))
	payload.SysUptimeStart = payload.SysUptimeEnd - uint32(randomNum(offsetCenter, startOffset))

	payload.SrcIP = IPtoUint32(srcIp)
	payload.DstIP = IPtoUint32(dstIp)

	if nextHopIp != "" {
		payload.NextHopIP = IPtoUint32(nextHopIp)
	}

	payload.SrcPort = srcPort
	payload.DstPort = dstPort

	payload.NumOctets = uint32(bytes)

	return *payload
}

// patch up the common fields of the packets
func FillCommonFields(
	payload *NetflowPayload,
	numPktOct int,
	ipProtocol int,
	srcPrefixMask int) NetflowPayload {

	// Fill template with values not filled by caller
	// payload.SrcIP = IPtoUint32("10.154.20.12")
	// payload.DstIP = IPtoUint32("77.12.190.94")
	// payload.NextHopIP = IPtoUint32("150.20.145.1")
	// payload.SrcPort = uint16(9010)
	// payload.DstPort = uint16(MYSQL_PORT)
	// payload.SnmpInIndex = genRandUint16(UINT16_MAX)
	// payload.SnmpOutIndex = genRandUint16(UINT16_MAX)
	payload.NumPackets = genRandUint32(numPktOct)
	payload.NumOctets = genRandUint32(numPktOct)
	// payload.SysUptimeStart = rand.Uint32()
	// payload.SysUptimeEnd = rand.Uint32()
	payload.Padding1 = 0
	payload.IpProtocol = uint8(ipProtocol)
	payload.IpTos = 0
	payload.SrcAsNumber = genRandUint16(UINT16_MAX, nil)
	payload.DstAsNumber = genRandUint16(UINT16_MAX, nil)

	payload.SrcPrefixMask = uint8(srcPrefixMask)
	payload.DstPrefixMask = uint8(rand.Intn(32))
	payload.Padding2 = 0

	// now handle computed values
	if payload.SrcIP > payload.DstIP { // false-index
		payload.SnmpInIndex = 1
		payload.SnmpOutIndex = 2
	} else {
		payload.SnmpInIndex = 2
		payload.SnmpOutIndex = 1
	}

	uptime := int(sysUptime)
	payload.SysUptimeEnd = uint32(uptime - randomNum(10, 500))
	payload.SysUptimeStart = payload.SysUptimeEnd - uint32(randomNum(10, 500))

	// log.Infof("S&D : %x %x %d, %d", payload.SrcIP, payload.DstIP, payload.DstPort, payload.SnmpInIndex)
	// log.Infof("Time: %d %d %d", sysUptime, payload.SysUptimeStart, payload.SysUptimeEnd)

	return *payload
}

func genRandUint16(max int, randGen *rand.Rand) uint16 {
	if randGen == nil {
		return uint16(rand.Intn(max))
	} else {
		return uint16(randGen.Intn(max))
	}
}

func IPtoUint32(s string) uint32 {
	ip := net.ParseIP(s)
	return binary.BigEndian.Uint32(ip.To4())
}

func genRandUint32(max int) uint32 {
	return uint32(rand.Intn(max))
}
