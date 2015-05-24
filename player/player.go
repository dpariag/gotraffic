package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"net"
	"time"
	"fmt"
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
	flow   *flow.Flow           // Flow being played
	bridge network.BridgeGroup  // Interface that packets are written to
	ips    []net.IP             // IPs that flows will be played from
	index  uint64               // Index of the ip currently in use
	in     chan gopacket.Packet // channel that packets are returned on
	stats  stats.PlayerStats	// Rx and Tx stats
	stop   bool                 // Stop replay after the current packet
}

func NewPlayer(bridge network.BridgeGroup, f *flow.Flow, ips []net.IP) Player {
	p := player{in: make(chan gopacket.Packet, f.NumPackets()),
		bridge: bridge, flow:f, ips: ips}
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
func (fp *player) register(subIP net.IP) {
	firstPkt, err := newPacket(fp.flow.Packets()[0], rewriteSource, subIP)
	if err != nil {
		panic(err)
	}
	fp.bridge.Register(firstPkt.NetworkLayer().NetworkFlow().Src(), fp.in)
}

func (fp *player) play() {
	// Choose an IP to play packets from
	subIP := fp.ips[fp.index]
	fp.index = (fp.index + 1) % uint64(len(fp.ips))

	fp.stats.FlowsStarted++
	client := fp.flow.Client()
	// Write packets to the interface, respecting inter-packet gaps
	for _, p := range fp.flow.Packets() {
		// Clone the packet, re-writing the subscriber IP and
		// send it to correct side of the bridge group
		if p.Packet.NetworkLayer().NetworkFlow().Src() == client {
			newPkt, err := newPacket(p, rewriteSource, subIP)
			if err != nil {
				panic(err)
			}
			fp.bridge.SendClientPacket(newPkt)
		} else {
			newPkt, err := newPacket(p, rewriteDest, subIP)
			if err != nil {
				panic(err)
			}
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


func printPacket(prefix string, p gopacket.Packet) {
	src, dst := p.NetworkLayer().NetworkFlow().Endpoints()
	fmt.Println(prefix, src, " --> ", dst)
}

func printFlow(flow *flow.Flow) {
	for _, p := range flow.Packets() {
		printPacket("", p)
	}
}

