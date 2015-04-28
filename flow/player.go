package flow

import (
	//"fmt"
	"net"
	"time"
	"github.com/google/gopacket"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
)

type Player struct {
	flow	Flow					// Flow being played
	bridge	network.BridgeGroup		// Interface that packets are written to 
	in		chan gopacket.Packet	// channel that packets are returned on 
	stats	stats.Directional		// Rx and Tx stats
	done	chan *Player			// Written on completion (allows easy replay)
}

func NewPlayer(bridge network.BridgeGroup, f *Flow) *Player {
	p := Player{in:make(chan gopacket.Packet, len(f.pkts)),
				bridge:bridge, flow:*f}
	return &p
}

func (fp *Player) readPackets() {
	// TODO: Need timeout
	for i := 0; i < len(fp.flow.pkts); i++ {
		pkt := <-fp.in // read back the packet
		fp.stats.Rx.Packets++
		fp.stats.Rx.Bytes += uint64(pkt.Metadata().CaptureInfo.CaptureLength)
	}
}

func (fp *Player) register(srcIP net.IP) {
	firstPkt, err := newPacket(fp.flow.pkts[0], srcIP)
	if err != nil {
		panic(err)
	}
	fp.bridge.Register(firstPkt.NetworkLayer().NetworkFlow().Src(), fp.in)
}

func (fp *Player) Play(srcIP net.IP) {
	// Register for returned packets
	fp.register(srcIP)
	go fp.readPackets()

	// Write packets to the interface, respecting inter-packet gaps
	//start := time.Now()
	//fmt.Printf("Starting...playing %v pkts from %v\n", len(fp.flow.pkts), srcIP.String())
	for _, p := range fp.flow.pkts {
		// Clone the packet, re-writing the source IP and send it to correct side of the bridge group
		newPkt,err := newPacket(p, srcIP)
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
	//fmt.Printf("Done flow playback. Elapsed: %v\n", time.Now().Sub(start))
}

// Play a flow, and signal completion by writing fp to the given channel upon completion
// This allows the channel owner to easily restart the flow
func (fp *Player) Replay(srcIP net.IP, done chan *Player) {
	fp.Play(srcIP)
	// TODO: Wait for all packets to be read back before signaling completion
	done <- fp
}

func (fp *Player) Stats() stats.Directional {
	return fp.stats
}
