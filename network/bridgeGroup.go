package network

import (
	"github.com/google/gopacket"
	"time"
)

//TODO: These stats structs are reusable. Move into their own package

type Stats struct {
	Bytes   uint64
	Packets uint64
}

type DirectionalStats struct {
	Rx Stats
	Tx Stats
}

type BridgeGroupStats struct {
	Client DirectionalStats
	Server DirectionalStats
}

type BridgeGroup interface {
	Register(hash uint64, c chan gopacket.Packet)
	Deregister([]gopacket.Flow)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	Stats() BridgeGroupStats
	String() string
	Shutdown(timeout time.Duration)
}
