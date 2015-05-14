package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"net"
	"time"
)

type Player interface {
	// Play a flow exactly once. Returns when replay is complete
	PlayOnce()
	// Play a flow in a loop. Returns immediately
	Play()
	// Stop flow replay immediately
	Stop()
	//Pause()
	Stats() stats.PlayerStats
}

type player struct {
	flow   flow.Flow            // Flow being played
	bridge network.BridgeGroup  // Interface that packets are written to
	ips    []net.IP             // IPs that flows will be played from
	index  uint64               // Index of the ip currently in use
	in     chan gopacket.Packet // channel that packets are returned on
	stats  stats.PlayerStats	// Rx and Tx stats
	stop   bool                 // Stop replay after the current packet
}

func NewPlayer(bridge network.BridgeGroup, f *flow.Flow, ips []net.IP) Player {
	p := player{in: make(chan gopacket.Packet, f.NumPackets()),
		bridge: bridge, flow: *f, ips: ips}
	for _, ip := range ips {
		p.register(ip)
	}
	go p.readPackets()
	return &p
}

//TODO: This does not work when a flow is replayed more than once
func (fp *player) readPackets() {
	for fp.stop == false || fp.stats.Rx.Packets < fp.stats.Tx.Packets {
		pkt := <-fp.in // read back the packet
		fp.stats.Rx.Packets++
		fp.stats.Rx.Bytes += uint64(pkt.Metadata().CaptureInfo.CaptureLength)
	}
}

// Registration is a bit of a hack
func (fp *player) register(srcIP net.IP) {
	firstPkt, err := newPacket(fp.flow.Packets()[0], srcIP)
	if err != nil {
		panic(err)
	}
	fp.bridge.Register(firstPkt.NetworkLayer().NetworkFlow().Src(), fp.in)
}

func (fp *player) play() {
	// Choose an IP to play packets from
	srcIP := fp.ips[fp.index]
	fp.index = (fp.index + 1) % uint64(len(fp.ips))

	fp.stats.FlowsStarted++
	// Write packets to the interface, respecting inter-packet gaps
	for _, p := range fp.flow.Packets() {
		// Clone the packet, re-writing the source IP and
		// send it to correct side of the bridge group
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
		if fp.stop {
			return
		}
	}
	fp.stats.FlowsCompleted++
}

func (fp *player) Play() {
	go func() {
		for fp.stop != true {
			fp.play()
		}
	}()
}

func (fp *player) PlayOnce() {
	fp.play()
}

func (fp *player) Stop() {
	fp.stop = true
}

func (fp *player) Stats() stats.PlayerStats {
	return fp.stats
}
