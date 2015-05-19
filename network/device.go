package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
)

const (
	clientDevice = 1
	serverDevice = 2
)

type Device interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	Send(p gopacket.Packet)
	Stats() stats.Directional
	Shutdown(timeout time.Duration)
}

type BridgeGroup interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	Stats() stats.BridgeGroupStats
	Shutdown(timeout time.Duration)
}
