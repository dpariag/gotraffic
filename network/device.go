package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
)

type Device interface {
	Init()
	Register(hash uint64, c chan gopacket.Packet)
	Send(p *gopacket.Packet)
	Stats() stats.Directional
	Shutdown(timeout time.Duration)
}

type BridgeGroup interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	Deregister([]gopacket.Flow)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	Stats() stats.BridgeGroupStats
	String() string
	Shutdown(timeout time.Duration)
}
