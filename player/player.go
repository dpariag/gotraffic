package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"net"
	"time"
)

type Player struct {
	flow   flow.Flow            // Flow being played
	bridge network.BridgeGroup  // Interface that packets are written to
	ips    []net.IP             // IPs that flows will be played from
	index  uint64               // Index of the ip currently in use
	in     chan gopacket.Packet // channel that packets are returned on
	stats  stats.Directional    // Rx and Tx stats
	done   chan *Player         // Written on completion (allows easy replay)
}

func NewPlayer(bridge network.BridgeGroup, f *flow.Flow, ips []net.IP) *Player {
	p := Player{in: make(chan gopacket.Packet, f.NumPackets()),
		bridge: bridge, flow: *f, ips: ips}
	for _, ip := range ips {
		p.register(ip)
	}
	return &p
}

func (fp *Player) readPackets() {
	// TODO: Need timeout
	for i := uint64(0); i < fp.flow.NumPackets(); i++ {
		pkt := <-fp.in // read back the packet
		fp.stats.Rx.Packets++
		fp.stats.Rx.Bytes += uint64(pkt.Metadata().CaptureInfo.CaptureLength)
	}
}

// Registration is a bit of a hack
func (fp *Player) register(srcIP net.IP) {
	firstPkt, err := newPacket(fp.flow.Packets()[0], srcIP)
	if err != nil {
		panic(err)
	}
	fp.bridge.Register(firstPkt.NetworkLayer().NetworkFlow().Src(), fp.in)
}

func (fp *Player) Play() {
	// Choose an IP to play packets from
	srcIP := fp.ips[fp.index]
	fp.index = (fp.index + 1) % uint64(len(fp.ips))
	go fp.readPackets()

	// Write packets to the interface, respecting inter-packet gaps
	for _, p := range fp.flow.Packets() {
		// Clone the packet, re-writing the source IP and send it to correct side of the bridge group
		newPkt, err := newPacket(p, srcIP)
		if err != nil {
			panic(err)
		}
		if p.Packet.NetworkLayer().NetworkFlow().Src() == fp.flow.Client() {
			fp.bridge.SendClientPacket(newPkt)
		} else {
			fp.bridge.SendServerPacket(newPkt)
		}
		fp.stats.Tx.Packets++
		fp.stats.Tx.Bytes += uint64(p.Packet.Metadata().CaptureInfo.CaptureLength)
		time.Sleep(p.Gap)
	}
}

// Play a flow, and signal completion by writing fp to the given channel upon completion
// This allows the channel owner to easily restart the flow
func (fp *Player) Replay(done chan *Player) {
	fp.Play()
	// TODO: Wait for all packets to be read back before signaling completion
	done <- fp
}

func (fp *Player) Stats() stats.Directional {
	return fp.stats
}
