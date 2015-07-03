package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
)

type BridgeGroup interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	Stats() stats.BridgeGroupStats
	Shutdown(timeout time.Duration)
}

type bridgeGroup struct {
	client Device
	server Device
	stats  stats.BridgeGroupStats
}

func (bg *bridgeGroup) Register(ep gopacket.Endpoint, c chan gopacket.Packet) {
	bg.client.Register(ep, c)
	bg.server.Register(ep, c)
}

func (bg *bridgeGroup) SendClientPacket(p gopacket.Packet) {
	bg.client.Send(p)
}

func (bg *bridgeGroup) SendServerPacket(p gopacket.Packet) {
	bg.server.Send(p)
}

func (bg *bridgeGroup) Stats() stats.BridgeGroupStats {
	bg.stats.Client = bg.client.Stats()
	bg.stats.Server = bg.server.Stats()
	return bg.stats
}

func (bg *bridgeGroup) Shutdown(timeout time.Duration) {
	bg.client.Shutdown(timeout)
	bg.server.Shutdown(timeout)
}
