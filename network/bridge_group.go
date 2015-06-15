package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
)

type bridgeGroup struct {
	client Device
	server Device
	stats  stats.BridgeGroupStats
}

func NewBridgeGroup(client string, server string) *bridgeGroup {
	return &bridgeGroup{client: NewPCAPDevice(client, clientDevice),
		server: NewPCAPDevice(server, serverDevice)}
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
