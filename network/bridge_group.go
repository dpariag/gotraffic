package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
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
