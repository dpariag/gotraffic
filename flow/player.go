package flow

import (
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/network"
	"github.com/google/gopacket"
	"time"
)

type Player struct {
	flow	Flow					// Flow being played
	bridge	network.BridgeGroup		// Interface that packets are written to 
	in		chan gopacket.Packet	// channel that packets are returned on 
	stats	network.DirectionalStats
	done	chan *Player			// Written on completion (allows easy replay)
}

func NewPlayer(bridge network.BridgeGroup, f *Flow) *Player {
	p := Player{in:make(chan gopacket.Packet, len(f.pkts)),
				bridge:bridge, flow:*f}
	p.bridge.Register(f.Hash(), p.in)
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

func (fp *Player) Play() {
	// Read packets from input channel
	go fp.readPackets()

	// Write packets to the interface, respecting inter-packet gaps
	start := time.Now()
	fmt.Printf("Starting...flow has %v pkts\n", len(fp.flow.pkts))
	for _, p := range fp.flow.pkts {
		// Send the embedded packet to the correct half of the bridge group
		if p.Packet.NetworkLayer().NetworkFlow().Src() == fp.flow.Client() {
			fp.bridge.SendClientPacket(p.Packet)
		} else {
			fp.bridge.SendServerPacket(p.Packet)
		}
		fp.stats.Tx.Packets++
		fp.stats.Tx.Bytes += uint64(p.Packet.Metadata().CaptureInfo.CaptureLength)
		time.Sleep(p.Gap)
	}
	fmt.Printf("Done flow playback. Elapsed: %v\n", time.Now().Sub(start))
}

// Play a flow, and signal completion by writing fp to the given channel upon completion
// This allows the channel owner to easily restart the flow
func (fp *Player) Replay(done chan *Player) {
	fp.Play()
	// TODO: Wait for all packets to be read back before signaling completion
	done <- fp
}

func (fp *Player) Stats() network.DirectionalStats {
	return fp.stats
}
