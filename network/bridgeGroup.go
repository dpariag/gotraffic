package network

import (
	"time"
	"code.google.com/p/gopacket"
)

type Stats struct {
	Bytes  	uint64 
	Packets uint64
}

type DeviceStats struct {
	Rx Stats
	Tx Stats
}

type BridgeGroupStats struct {
	client DeviceStats
	server DeviceStats
}

type BridgeGroup interface {
	Register(flows []gopacket.Flow, c chan gopacket.Packet)
	Deregister([]gopacket.Flow)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	GetStats() (BridgeGroupStats)
	String() string
	Shutdown(timeout time.Duration)
}
