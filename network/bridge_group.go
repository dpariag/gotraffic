package network

import (
	"time"
	"github.com/google/gopacket"
	"git.svc.rocks/dpariag/gotraffic/stats"
)

type BridgeGroupStats struct {
	Client stats.Directional
	Server stats.Directional
}

type BridgeGroup interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	Deregister([]gopacket.Flow)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	Stats() BridgeGroupStats
	String() string
	Shutdown(timeout time.Duration)
}
