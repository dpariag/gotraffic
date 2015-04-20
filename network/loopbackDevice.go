// A conceptual loopback bridge group, intended primarily for unit testing
// Packets written to this bridge group are immediately returned to the interface
// Registered listeners can be notified immediately
// Requires no underlying physical device 

package network

import (
	"github.com/google/gopacket"
	"time"
)

type LoopbackBridgeGroup struct {
	channels map[uint64] chan gopacket.Packet
	stats BridgeGroupStats
}

func NewLoopbackBridgeGroup() *LoopbackBridgeGroup {
	i := LoopbackBridgeGroup{}
	i.channels = make(map[uint64]chan gopacket.Packet)
	return &i
}

// Register flows with the interface.
func (l *LoopbackBridgeGroup) Register(hash uint64, c chan gopacket.Packet) {
	l.channels[hash] = c
}

func (l *LoopbackBridgeGroup) Deregister(flows []gopacket.Flow) {
}

func (l *LoopbackBridgeGroup) SendClientPacket(p gopacket.Packet) {
	l.stats.Client.Tx.Packets++
	l.stats.Client.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	l.stats.Server.Rx.Packets++
	l.stats.Server.Rx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	ch := l.channels[p.NetworkLayer().NetworkFlow().FastHash()]
	if ch != nil {
		ch <- p
	}
}

func (l *LoopbackBridgeGroup) SendServerPacket(p gopacket.Packet) {
	l.stats.Server.Tx.Packets++
	l.stats.Server.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	l.stats.Client.Rx.Packets++
	l.stats.Client.Rx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	ch := l.channels[p.NetworkLayer().NetworkFlow().FastHash()]
	if ch != nil {
		ch <- p
	}
}

func (l *LoopbackBridgeGroup) Shutdown(timeout time.Duration) {}

func (l *LoopbackBridgeGroup) Stats() BridgeGroupStats {
	return l.stats
}

func (l *LoopbackBridgeGroup) String() string {
	return "loopbackBridgeGroup"
}
