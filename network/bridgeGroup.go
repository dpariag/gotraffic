package network

import (
	"time"
	"github.com/google/gopacket"
	"git.svc.rocks/dpariag/gotraffic/stats"
)
/*
type Stats struct {
	Bytes   uint64
	Packets uint64
}

type DirectionalStats struct {
	Rx Stats
	Tx Stats
}
*/

type BridgeGroupStats struct {
	Client stats.Directional
	Server stats.Directional
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
