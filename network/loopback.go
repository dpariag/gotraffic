// A conceptual loopback bridge group, intended primarily for unit testing
// Packets written to this bridge group are immediately returned to the interface
// Registered listeners can be notified immediately
// Requires no underlying physical device

package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
)

type LoopbackBridgeGroup struct {
	channels map[uint64]chan gopacket.Packet
	stats    stats.BridgeGroupStats
}

func NewLoopbackBridgeGroup() *LoopbackBridgeGroup {
	return &LoopbackBridgeGroup{channels: make(map[uint64]chan gopacket.Packet)}
}

// Register an endpoint with the interface.
func (l *LoopbackBridgeGroup) Register(ep gopacket.Endpoint, c chan gopacket.Packet) {
	l.channels[ep.FastHash()] = c
}

func (l *LoopbackBridgeGroup) SendClientPacket(p gopacket.Packet) {
	l.stats.Client.Tx.Packets++
	l.stats.Client.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	l.stats.Server.Rx.Packets++
	l.stats.Server.Rx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	ch := l.channels[p.NetworkLayer().NetworkFlow().Src().FastHash()]
	if ch != nil {
		ch <- p
	}
}

func (l *LoopbackBridgeGroup) SendServerPacket(p gopacket.Packet) {
	l.stats.Server.Tx.Packets++
	l.stats.Server.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	l.stats.Client.Rx.Packets++
	l.stats.Client.Rx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	ch := l.channels[p.NetworkLayer().NetworkFlow().Dst().FastHash()]
	if ch != nil {
		ch <- p
	}
}

func (l *LoopbackBridgeGroup) Shutdown(timeout time.Duration) {
    start := time.Now()
    for _, ch := range l.channels {
        for len(ch) > 0 && time.Since(start) < timeout {
            time.Sleep(1 * time.Millisecond)
        }
    }
}

func (l *LoopbackBridgeGroup) Stats() stats.BridgeGroupStats {
	return l.stats
}
