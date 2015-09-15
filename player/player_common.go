package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
)

const (
	stopped = 1 // No more packets to send
	playing = 2 // Playing packets
)

type Player interface {
	PlayOnce()	// Play a flow exactly once. Returns when replay is complete
	Play()		// Play a flow in a loop. Returns immediately
	Stop()		// Stop flow replay immediately
	Stats() stats.PlayerStats
}

// Things that are expected to be common to any struct that implements the Player interface
// Embed and specialize.
type playerCommon struct {
	flow   *flow.Flow           // Flow being played
	bridge network.BridgeGroup  // Bridge group that packets are written to
	in     chan gopacket.Packet // channel that packets are returned on
	stats  stats.PlayerStats    // Rx and Tx stats
	state  int                  // State of the player
}

func newPlayerCommon(bridge network.BridgeGroup, f *flow.Flow) playerCommon {
    return playerCommon{flow: f, bridge: bridge, state: stopped,
        in: make(chan gopacket.Packet, f.NumPackets())}
}

//TODO: This does not work when a flow is replayed more than once
func (p *playerCommon) readPackets() {
	for true {
		pkt := <-p.in // read back the packet
		p.stats.Rx.Packets++
		p.stats.Rx.Bytes += uint64(pkt.Metadata().CaptureInfo.CaptureLength)
	}
}

func (p *playerCommon) Stats() stats.PlayerStats {
	return p.stats
}
